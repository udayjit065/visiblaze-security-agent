package handlers

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func HealthHandler(ctx context.Context, request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	body := `{"status":"ok","time":"` + time.Now().UTC().Format(time.RFC3339) + `"}`

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       body,
	}, nil
}
