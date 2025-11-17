package main

import (
	"context"
	"log"
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

func handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Debug logging
	log.Printf("Request received: Method=%s, RawPath=%s, Path=%s", request.RequestContext.HTTP.Method, request.RawPath, request.RequestContext.HTTP.Path)

	// Strip stage prefix from path for routing
	path := request.RawPath
	if stage := request.RequestContext.Stage; stage != "" && stage != "$default" {
		stagePrefix := "/" + stage
		if len(path) > len(stagePrefix) && path[:len(stagePrefix)] == stagePrefix {
			path = path[len(stagePrefix):]
		}
	}
	log.Printf("Stripped path: %s", path)

	// CORS headers
	headers := map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type,X-API-Key",
	}

	// Handle OPTIONS
	if request.RequestContext.HTTP.Method == "OPTIONS" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers:    headers,
		}, nil
	}

	// Validate API key for write operations
	if request.RequestContext.HTTP.Method == "POST" {
		log.Printf("Validating API key for POST request")
		log.Printf("Available headers: %v", request.Headers)
		key := request.Headers["x-api-key"]
		log.Printf("Received API key: %s", key)
		log.Printf("Expected API key: %s", apiKey)
		if key != apiKey {
			log.Printf("API key validation failed")
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 401,
				Headers:    headers,
				Body:       `{"error":"Unauthorized"}`,
			}, nil
		}
		log.Printf("API key validation passed")
	}

	// Route
	switch {
	case request.RequestContext.HTTP.Method == "POST" && (path == "/ingest" || path == "/ingest/"):
		return handlers.IngestHandler(ctx, request, dynamoClient, headers)
	case request.RequestContext.HTTP.Method == "GET" && path == "/hosts":
		return handlers.HostsListHandler(ctx, request, dynamoClient, headers)
	case request.RequestContext.HTTP.Method == "GET" && request.PathParameters["hostId"] != "":
		return handlers.HostDetailHandler(ctx, request, dynamoClient, headers)
	case request.RequestContext.HTTP.Method == "GET" && path == "/apps":
		return handlers.PackagesHandler(ctx, request, dynamoClient, headers)
	case request.RequestContext.HTTP.Method == "GET" && path == "/cis-results":
		return handlers.CISResultsHandler(ctx, request, dynamoClient, headers)
	case request.RequestContext.HTTP.Method == "GET" && path == "/health":
		return handlers.HealthHandler(ctx, request, headers)
	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 404,
			Headers:    headers,
			Body:       `{"error":"Not found"}`,
		}, nil
	}
}

func main() {
	lambda.Start(handler)
}
