package handlers

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HealthHandler(ctx context.Context, request events.APIGatewayV2HTTPRequest, headers map[string]string) (events.APIGatewayV2HTTPResponse, error) {
	body := `{"status":"ok","time":"` + time.Now().UTC().Format(time.RFC3339) + `"}`

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       body,
	}, nil
}
