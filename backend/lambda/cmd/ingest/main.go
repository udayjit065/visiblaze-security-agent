package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/visiblaze/sec-agent/backend/lambda/internal/handlers"
)

var (
	dynamoClient *dynamodb.Client
	apiKey       string
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.Background())
	dynamoClient = dynamodb.NewFromConfig(cfg)
	apiKey = os.Getenv("API_KEY")
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// CORS headers
	headers := map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type,X-API-Key",
	}

	// Handle OPTIONS
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
		}, nil
	}

	// Validate API key for write operations
	if request.HTTPMethod == "POST" {
		key := request.Headers["X-API-Key"]
		if key != apiKey {
			return events.APIGatewayProxyResponse{
				StatusCode: 401,
				Headers:    headers,
				Body:       `{"error":"Unauthorized"}`,
			}, nil
		}
	}

	// Route
	switch {
	case request.HTTPMethod == "POST" && (request.Path == "/ingest" || request.Path == "/ingest/"):
		return handlers.IngestHandler(ctx, request, dynamoClient, headers)
	case request.HTTPMethod == "GET" && request.Path == "/hosts":
		return handlers.HostsListHandler(ctx, request, dynamoClient, headers)
	case request.HTTPMethod == "GET" && request.PathParameters["hostId"] != "":
		return handlers.HostDetailHandler(ctx, request, dynamoClient, headers)
	case request.HTTPMethod == "GET" && request.Path == "/apps":
		return handlers.PackagesHandler(ctx, request, dynamoClient, headers)
	case request.HTTPMethod == "GET" && request.Path == "/cis-results":
		return handlers.CISResultsHandler(ctx, request, dynamoClient, headers)
	case request.HTTPMethod == "GET" && request.Path == "/health":
		return handlers.HealthHandler(ctx, request, headers)
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Headers:    headers,
			Body:       `{"error":"Not found"}`,
		}, nil
	}
}

func main() {
	lambda.Start(handler)
}
