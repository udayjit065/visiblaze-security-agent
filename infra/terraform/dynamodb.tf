# Hosts Table
resource "aws_dynamodb_table" "hosts" {
  name           = "vis_hosts"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "host_id"
  
  attribute {
    name = "host_id"
    type = "S"
  }

  attribute {
    name = "last_seen"
    type = "S"
  }

  global_secondary_index {
    name            = "LastSeenIndex"
    hash_key        = "last_seen"
    projection_type = "ALL"
  }

  ttl {
    attribute_name = "expiration"
    enabled        = false
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Table = "hosts"
  }
}

# Packages Table
resource "aws_dynamodb_table" "packages" {
  name           = "vis_packages"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "host_id"
  range_key      = "pkg_key"

  attribute {
    name = "host_id"
    type = "S"
  }

  attribute {
    name = "pkg_key"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Table = "packages"
  }
}

# CIS Results Table
resource "aws_dynamodb_table" "cis_results" {
  name           = "vis_cis_results"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "host_id"
  range_key      = "check_id"

  attribute {
    name = "host_id"
    type = "S"
  }

  attribute {
    name = "check_id"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Table = "cis_results"
  }
}
