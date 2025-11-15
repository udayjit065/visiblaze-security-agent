package handlers

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/visiblaze/sec-agent/backend/lambda/internal/models"
)

func attrString(attr types.AttributeValue) string {
	if s, ok := attr.(*types.AttributeValueMemberS); ok {
		return s.Value
	}
	return ""
}

func attrStringSlice(attr types.AttributeValue) []string {
	if ss, ok := attr.(*types.AttributeValueMemberSS); ok {
		return ss.Value
	}
	if s, ok := attr.(*types.AttributeValueMemberS); ok && s.Value != "" {
		return []string{s.Value}
	}
	return []string{}
}

func HostsListHandler(ctx context.Context, req events.APIGatewayProxyRequest,
	client *dynamodb.Client, headers map[string]string) (events.APIGatewayProxyResponse, error) {

	out, _ := client.Scan(ctx, &dynamodb.ScanInput{
		TableName: str("vis_hosts"),
		Limit:     int32Ptr(100),
	})

	hosts := []models.Host{}
	for _, item := range out.Items {
		h := models.Host{
			HostID:       attrString(item["host_id"]),
			Hostname:     attrString(item["hostname"]),
			OSID:         attrString(item["os_id"]),
			OSVersion:    attrString(item["os_version"]),
			Kernel:       attrString(item["kernel"]),
			IPAddresses:  attrStringSlice(item["ip_addresses"]),
			AgentVersion: attrString(item["agent_version"]),
		}
		hosts = append(hosts, h)
	}

	body, _ := json.Marshal(map[string]interface{}{"hosts": hosts})
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func HostDetailHandler(ctx context.Context, req events.APIGatewayProxyRequest,
	client *dynamodb.Client, headers map[string]string) (events.APIGatewayProxyResponse, error) {

	hostID := req.PathParameters["hostId"]

	hostOut, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: str("vis_hosts"),
		Key: map[string]types.AttributeValue{
			"host_id": &types.AttributeValueMemberS{Value: hostID},
		},
	})
	if err != nil || hostOut.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers:    headers,
			Body:       `{"error":"host not found"}`,
		}, nil
	}

	host := models.Host{
		HostID:       attrString(hostOut.Item["host_id"]),
		Hostname:     attrString(hostOut.Item["hostname"]),
		OSID:         attrString(hostOut.Item["os_id"]),
		OSVersion:    attrString(hostOut.Item["os_version"]),
		Kernel:       attrString(hostOut.Item["kernel"]),
		IPAddresses:  attrStringSlice(hostOut.Item["ip_addresses"]),
		AgentVersion: attrString(hostOut.Item["agent_version"]),
	}

	cisOut, _ := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              str("vis_cis_results"),
		KeyConditionExpression: str("host_id = :hostId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hostId": &types.AttributeValueMemberS{Value: hostID},
		},
	})
	cis := []models.CISResult{}
	if cisOut != nil {
		for _, item := range cisOut.Items {
			evJSON := attrString(item["evidence"])
			var evidence map[string]interface{}
			json.Unmarshal([]byte(evJSON), &evidence)
			if evidence == nil {
				evidence = map[string]interface{}{}
			}
			cis = append(cis, models.CISResult{
				CheckID:   attrString(item["check_id"]),
				Title:     attrString(item["title"]),
				Status:    attrString(item["status"]),
				Evidence:  evidence,
				Timestamp: attrString(item["last_ts"]),
			})
		}
	}

	pkgOut, _ := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              str("vis_packages"),
		KeyConditionExpression: str("host_id = :hostId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hostId": &types.AttributeValueMemberS{Value: hostID},
		},
	})
	packages := []models.Package{}
	if pkgOut != nil {
		for _, item := range pkgOut.Items {
			packages = append(packages, models.Package{
				Name:    attrString(item["name"]),
				Version: attrString(item["version"]),
				Arch:    attrString(item["arch"]),
				Manager: attrString(item["manager"]),
				Source:  attrString(item["source"]),
			})
		}
	}

	body, _ := json.Marshal(map[string]interface{}{
		"host":        host,
		"cis_results": cis,
		"packages":    packages,
	})
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func PackagesHandler(ctx context.Context, req events.APIGatewayProxyRequest,
	client *dynamodb.Client, headers map[string]string) (events.APIGatewayProxyResponse, error) {

	out, _ := client.Scan(ctx, &dynamodb.ScanInput{
		TableName: str("vis_packages"),
		Limit:     int32Ptr(1000),
	})

	packages := []models.Package{}
	for _, item := range out.Items {
		packages = append(packages, models.Package{
			Name:    attrString(item["name"]),
			Version: attrString(item["version"]),
			Arch:    attrString(item["arch"]),
			Manager: attrString(item["manager"]),
			Source:  attrString(item["source"]),
		})
	}

	body, _ := json.Marshal(map[string]interface{}{"packages": packages})
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func CISResultsHandler(ctx context.Context, req events.APIGatewayProxyRequest,
	client *dynamodb.Client, headers map[string]string) (events.APIGatewayProxyResponse, error) {

	out, _ := client.Scan(ctx, &dynamodb.ScanInput{
		TableName: str("vis_cis_results"),
		Limit:     int32Ptr(1000),
	})

	results := []models.CISResult{}
	for _, item := range out.Items {
		evJSON := attrString(item["evidence"])
		var evidence map[string]interface{}
		json.Unmarshal([]byte(evJSON), &evidence)
		if evidence == nil {
			evidence = map[string]interface{}{}
		}
		results = append(results, models.CISResult{
			CheckID:   attrString(item["check_id"]),
			Title:     attrString(item["title"]),
			Status:    attrString(item["status"]),
			Evidence:  evidence,
			Timestamp: attrString(item["last_ts"]),
		})
	}

	body, _ := json.Marshal(map[string]interface{}{"cis_results": results})
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func int32Ptr(i int) *int32 {
	v := int32(i)
	return &v
}
