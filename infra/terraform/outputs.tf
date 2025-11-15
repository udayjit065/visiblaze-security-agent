output "api_endpoint" {
  description = "API Gateway endpoint URL"
  value       = aws_apigatewayv2_stage.prod.invoke_url
}

output "api_key" {
  description = "API key for agent authentication"
  value       = random_password.api_key.result
  sensitive   = true
}

output "api_key_ssm_parameter" {
  description = "SSM Parameter path for API key"
  value       = aws_ssm_parameter.api_key.name
}

output "dynamodb_hosts_table" {
  description = "Hosts table name"
  value       = aws_dynamodb_table.hosts.name
}

output "dynamodb_packages_table" {
  description = "Packages table name"
  value       = aws_dynamodb_table.packages.name
}

output "dynamodb_cis_results_table" {
  description = "CIS results table name"
  value       = aws_dynamodb_table.cis_results.name
}

output "lambda_function_name" {
  description = "Lambda function name"
  value       = aws_lambda_function.ingest.function_name
}
