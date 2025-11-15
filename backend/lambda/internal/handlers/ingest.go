package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/visiblaze/sec-agent/backend/lambda/internal/models"
)

func IngestHandler(ctx context.Context, req events.APIGatewayProxyRequest,
	client *dynamodb.Client, headers map[string]string) (events.APIGatewayProxyResponse, error) {

	var payload models.IngestPayload
	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers:    headers,
			Body:       fmt.Sprintf(`{"error":"Invalid JSON: %s"}`, err.Error()),
		}, nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	exprValues := map[string]types.AttributeValue{
		":hostname":   &types.AttributeValueMemberS{Value: payload.Host.Hostname},
		":os_id":      &types.AttributeValueMemberS{Value: payload.Host.OSID},
		":os_version": &types.AttributeValueMemberS{Value: payload.Host.OSVersion},
		":kernel":     &types.AttributeValueMemberS{Value: payload.Host.Kernel},
		":agent_ver":  &types.AttributeValueMemberS{Value: payload.Host.AgentVersion},
		":last_seen":  &types.AttributeValueMemberS{Value: now},
		":first_seen": &types.AttributeValueMemberS{Value: now},
	}

	updateExpr := "SET hostname = :hostname, os_id = :os_id, os_version = :os_version, kernel = :kernel, agent_version = :agent_ver, last_seen = :last_seen, first_seen = if_not_exists(first_seen, :first_seen)"

	var removeClause string
	ipSet := uniqueStrings(payload.Host.IPAddresses)
	if len(ipSet) > 0 {
		exprValues[":ip_addresses"] = &types.AttributeValueMemberSS{Value: ipSet}
		updateExpr += ", ip_addresses = :ip_addresses"
	} else {
		removeClause = " REMOVE ip_addresses"
	}

	if removeClause != "" {
		updateExpr += removeClause
	}

	_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 str("vis_hosts"),
		Key:                       map[string]types.AttributeValue{"host_id": &types.AttributeValueMemberS{Value: payload.Host.HostID}},
		UpdateExpression:          str(updateExpr),
		ExpressionAttributeValues: exprValues,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       fmt.Sprintf(`{"error":"Failed to store host: %s"}`, err.Error()),
		}, nil
	}

	// Delete existing packages for this host
	scanOut, _ := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        str("vis_packages"),
		FilterExpression: str("host_id = :hostId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hostId": &types.AttributeValueMemberS{Value: payload.Host.HostID},
		},
	})

	for _, item := range scanOut.Items {
		hostIDAttr := item["host_id"].(*types.AttributeValueMemberS)
		pkgKeyAttr := item["pkg_key"].(*types.AttributeValueMemberS)
		client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: str("vis_packages"),
			Key: map[string]types.AttributeValue{
				"host_id": &types.AttributeValueMemberS{Value: hostIDAttr.Value},
				"pkg_key": &types.AttributeValueMemberS{Value: pkgKeyAttr.Value},
			},
		})
	}

	// Insert packages
	for _, pkg := range payload.Packages {
		pkgKey := fmt.Sprintf("%s#%s", pkg.Name, pkg.Arch)
		pkgItem := map[string]types.AttributeValue{
			"host_id": &types.AttributeValueMemberS{Value: payload.Host.HostID},
			"pkg_key": &types.AttributeValueMemberS{Value: pkgKey},
			"name":    &types.AttributeValueMemberS{Value: pkg.Name},
			"version": &types.AttributeValueMemberS{Value: pkg.Version},
			"arch":    &types.AttributeValueMemberS{Value: pkg.Arch},
			"manager": &types.AttributeValueMemberS{Value: pkg.Manager},
			"source":  &types.AttributeValueMemberS{Value: pkg.Source},
		}
		client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: str("vis_packages"),
			Item:      pkgItem,
		})
	}

	// Upsert CIS results (latest only per check_id)
	for _, result := range payload.CISResults {
		evJSON, _ := json.Marshal(result.Evidence)
		cisItem := map[string]types.AttributeValue{
			"host_id":  &types.AttributeValueMemberS{Value: payload.Host.HostID},
			"check_id": &types.AttributeValueMemberS{Value: result.CheckID},
			"title":    &types.AttributeValueMemberS{Value: result.Title},
			"status":   &types.AttributeValueMemberS{Value: result.Status},
			"evidence": &types.AttributeValueMemberS{Value: string(evJSON)},
			"last_ts":  &types.AttributeValueMemberS{Value: result.Timestamp},
		}
		client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: str("vis_cis_results"),
			Item:      cisItem,
		})
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       `{"status":"ok"}`,
	}, nil
}

func str(s string) *string {
	return &s
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}
