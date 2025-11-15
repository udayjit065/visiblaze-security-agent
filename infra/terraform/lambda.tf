# Package lambda binary into a zip that Terraform can deploy
data "archive_file" "lambda_package" {
  type        = "zip"
  source_file = "${path.module}/../../backend/lambda/bootstrap"
  output_path = "${path.module}/lambda_function.zip"
}

# Lambda function (uses Go custom runtime with bootstrap binary)
resource "aws_lambda_function" "ingest" {
  filename      = data.archive_file.lambda_package.output_path
  function_name = "${local.project_name}-ingest"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"
  timeout       = 60
  memory_size   = 256

  environment {
    variables = {
      HOSTS_TABLE       = aws_dynamodb_table.hosts.name
      PACKAGES_TABLE    = aws_dynamodb_table.packages.name
      CIS_RESULTS_TABLE = aws_dynamodb_table.cis_results.name
      API_KEY           = random_password.api_key.result
      ENVIRONMENT       = local.stage
    }
  }

  source_code_hash = data.archive_file.lambda_package.output_base64sha256

  depends_on = [aws_iam_role_policy.dynamodb_policy]
}

# Lambda permission for API Gateway
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.ingest.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}
