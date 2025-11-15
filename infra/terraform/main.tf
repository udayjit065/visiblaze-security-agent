locals {
  project_name = "visiblaze"
  stage        = "prod"
}

# Generate random API key
resource "random_password" "api_key" {
  length  = 32
  special = true
}

# Store API key in SSM Parameter Store
resource "aws_ssm_parameter" "api_key" {
  name  = "/${local.project_name}/api-key"
  type  = "SecureString"
  value = random_password.api_key.result

  tags = {
    Description = "API key for visiblaze agent"
  }
}
