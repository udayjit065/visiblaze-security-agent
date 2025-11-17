# HTTP API
resource "aws_apigatewayv2_api" "main" {
  name          = "${local.project_name}-api"
  protocol_type = "HTTP"
  api_key_selection_expression = "$request.header.x-api-key"  
  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers = ["Content-Type", "X-API-Key"]
    expose_headers = ["Content-Type"]
  }
}

# Stage
resource "aws_apigatewayv2_stage" "prod" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = local.stage
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_logs.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      resourcePath   = "$context.resourcePath"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }
}

# CloudWatch logs for API
resource "aws_cloudwatch_log_group" "api_logs" {
  name              = "/aws/apigateway/${local.project_name}"
  retention_in_days = 7
}

# Integration with Lambda
resource "aws_apigatewayv2_integration" "lambda" {
  api_id           = aws_apigatewayv2_api.main.id
  integration_type = "AWS_PROXY"
  integration_uri  = aws_lambda_function.ingest.arn
  payload_format_version = "2.0"  # <--- ADD THIS LINE
}
# Routes
resource "aws_apigatewayv2_route" "ingest" {
  api_id       = aws_apigatewayv2_api.main.id
  route_key    = "POST /ingest"
  target       = "integrations/${aws_apigatewayv2_integration.lambda.id}"
  authorization_type = "NONE"
  api_key_required   = false

  depends_on = [aws_apigatewayv2_integration.lambda]
}

resource "aws_apigatewayv2_route" "hosts_list" {
  api_id       = aws_apigatewayv2_api.main.id
  route_key    = "GET /hosts"
  target       = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_apigatewayv2_route" "host_detail" {
  api_id       = aws_apigatewayv2_api.main.id
  route_key    = "GET /hosts/{hostId}"
  target       = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_apigatewayv2_route" "packages" {
  api_id       = aws_apigatewayv2_api.main.id
  route_key    = "GET /apps"
  target       = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_apigatewayv2_route" "cis_results" {
  api_id       = aws_apigatewayv2_api.main.id
  route_key    = "GET /cis-results"
  target       = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_apigatewayv2_route" "health" {
  api_id       = aws_apigatewayv2_api.main.id
  route_key    = "GET /health"
  target       = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

# API Key for agent auth
