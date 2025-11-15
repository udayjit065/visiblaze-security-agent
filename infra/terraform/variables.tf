variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-south-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "visiblaze"
}

variable "api_key_length" {
  description = "Length of generated API key"
  type        = number
  default     = 32
}
