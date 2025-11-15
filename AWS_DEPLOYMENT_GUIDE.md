# Complete AWS Deployment Guide for Visiblaze Security Agent

This guide walks you through deploying Visiblaze Security Agent to AWS in real-time. You'll deploy:
- **Agent**: Linux daemon collecting security compliance data
- **Backend**: AWS Lambda API with DynamoDB storage
- **Frontend**: CloudFront-delivered React web dashboard
- **Infrastructure**: Terraform-managed VPC, IAM, API Gateway, Lambda, DynamoDB, CloudFront

## Prerequisites

1. **AWS Account** with:
   - AWS CLI v2 installed and configured with credentials
   - Permissions for EC2, Lambda, DynamoDB, API Gateway, CloudFront, IAM, S3

2. **Local Tools**:
   - Go 1.21+
   - Node 18+ npm
   - Terraform 1.5+ (download from terraform.io)
   - Docker (optional, for local DynamoDB testing)

3. **Verify AWS Credentials**:
```bash
aws sts get-caller-identity
# Should output your AWS account details
```

---

## Step 1: Build and Package the Agent Binary

### 1.1 Cross-compile Linux binary from Windows/macOS

```bash
cd d:/programming/dev/delta/projects/visiblaze-sec-agent
mkdir -p dist

cd agent
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags "-s -w -X main.Version=0.1.0" \
  -o ../dist/visiblaze-agent ./cmd/agent
```

**Output**: `dist/visiblaze-agent` — executable for Linux amd64

### 1.2 Create DEB/RPM packages (on Linux or WSL)

Install nfpm:
```bash
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
```

From repo root:
```bash
cd d:/programming/dev/delta/projects/visiblaze-sec-agent

# Create DEB
nfpm package -f packaging/nfpm.yaml -p deb -t dist/

# Create RPM (optional)
nfpm package -f packaging/nfpm.yaml -p rpm -t dist/
```

**Output**: 
- `dist/visiblaze-agent_0.1.0_amd64.deb`
- `dist/visiblaze-agent_0.1.0_amd64.rpm` (optional)

---

## Step 2: Build Lambda Handler and Deploy Infrastructure with Terraform

### 2.0 Build the Lambda handler binary

Terraform now packages the Lambda deployment archive automatically, but it expects a compiled `bootstrap` binary at `backend/lambda/bootstrap`. Use the helper script:

```bash
# From repo root
./scripts/build_lambda.sh
```

This cross-compiles the handler for Linux (`GOOS=linux GOARCH=amd64`) and places the executable where Terraform’s `archive_file` data source can find it.

### 2.1 Initialize Terraform

```bash
cd d:/programming/dev/delta/projects/visiblaze-sec-agent/infra/terraform

# Initialize Terraform (downloads providers)
terraform init

# List available variables
terraform variables
```

### 2.2 Create a `terraform.tfvars` file

Create `infra/terraform/terraform.tfvars`:
```hcl
aws_region              = "us-east-1"
project_name            = "visiblaze"
environment             = "prod"
lambda_memory           = 256
lambda_timeout          = 30
dynamodb_read_capacity  = 5
dynamodb_write_capacity = 5
enable_cloudfront       = true
```

### 2.3 Plan and Apply Terraform

```bash
# Preview changes
terraform plan -out=tfplan

# Apply (will create AWS resources)
terraform apply tfplan

# Save outputs
terraform output -json > outputs.json
cat outputs.json
# Note: api_gateway_invoke_url, cloudfront_domain_name, etc.
```

**Created Resources**:
- **DynamoDB Tables**: `vis_hosts`, `vis_packages`, `vis_cis_results`
- **Lambda Function**: visiblaze-ingest handler (binary zip deployed)
- **API Gateway**: REST API with routes (/ingest, /hosts, /apps, /cis-results, /health)
- **CloudFront**: CDN distribution pointing to S3 frontend bucket
- **IAM Roles**: Lambda execution role with DynamoDB permissions
- **SSM Parameter**: API key stored securely

---

## Step 3: Build and Deploy Frontend

### 3.1 Build React App

```bash
cd d:/programming/dev/delta/projects/visiblaze-sec-agent/web

# Install dependencies
npm install

# Build for production
VITE_API_BASE_URL=https://your-api-gateway-url npm run build
# Replace your-api-gateway-url with the invoke URL from Terraform outputs
```

**Output**: `dist/` directory with optimized assets

### 3.2 Upload to S3

```bash
# Get S3 bucket name from Terraform outputs
BUCKET=$(terraform output -raw frontend_bucket_name)

# Upload built assets
aws s3 sync dist/ s3://${BUCKET}/ \
  --delete \
  --cache-control "public, max-age=3600" \
  --region us-east-1

# Invalidate CloudFront cache
DISTRIBUTION=$(terraform output -raw cloudfront_distribution_id)
aws cloudfront create-invalidation \
  --distribution-id ${DISTRIBUTION} \
  --paths "/*"
```

**Output**: Frontend deployed at CloudFront URL (from Terraform outputs)

---

## Step 4: Launch Agent on EC2

### 4.1 Create EC2 Instance

```bash
aws ec2 run-instances \
  --image-ids ami-0c02fb55956c7d316 \
  --instance-type t3.micro \
  --key-name your-key-pair \
  --security-groups default \
  --region us-east-1 \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=visiblaze-agent-1}]'
```

Note the instance ID and public IP.

### 4.2 SSH into EC2 instance

```bash
ssh -i /path/to/key.pem ec2-user@<public-ip>
# or ubuntu@<public-ip> if using Ubuntu AMI
```

### 4.3 Install Agent from DEB/RPM

Download the package from your S3 bucket or copy from local:
```bash
# On EC2
wget https://s3.amazonaws.com/your-bucket/visiblaze-agent_0.1.0_amd64.deb
sudo apt update && sudo apt install -y ./visiblaze-agent_0.1.0_amd64.deb
```

Or manually copy binary:
```bash
sudo mkdir -p /usr/local/bin /var/lib/visiblaze-agent /var/log/visiblaze-agent
sudo cp visiblaze-agent /usr/bin/
sudo chown -R nobody:nobody /var/lib/visiblaze-agent /var/log/visiblaze-agent
```

### 4.4 Create Agent Config

Create `/etc/visiblaze-agent/config.yaml`:

```yaml
api_base_url: "https://your-api-gateway-url"  # from Terraform outputs
api_key: "your-api-key"  # from SSM Parameter Store or terraform outputs
collection_interval_minutes: 15
disable_ipv6_check: false
distro_hint: "ubuntu"  # or "rhel", "amazon-linux" etc.
```

Retrieve API key from SSM:
```bash
aws ssm get-parameter \
  --name /visiblaze/api-key \
  --with-decryption \
  --query 'Parameter.Value' \
  --output text \
  --region us-east-1
```

### 4.5 Start Agent Service

If using DEB package (systemd):
```bash
sudo systemctl enable visiblaze-agent
sudo systemctl start visiblaze-agent
sudo systemctl status visiblaze-agent
```

Check logs:
```bash
sudo tail -f /var/log/visiblaze-agent/agent.log
```

Or run agent manually (for debugging):
```bash
VISIBLAZE_LOG_DIR=/tmp/logs /usr/bin/visiblaze-agent \
  -config /etc/visiblaze-agent/config.yaml \
  -once
```

---

## Step 5: View Data in Real-Time

### 5.1 Open Frontend Dashboard

In your browser:
```
https://your-cloudfront-domain
# or
https://your-api-gateway-url (if frontend deployed there directly)
```

You should see:
- **Hosts List**: Showing the EC2 instance(s) running agents
- **Host Detail**: CIS check results, kernel version, IP addresses
- **Packages**: Installed packages aggregated across hosts
- **CIS Results**: Compliance check status (pass/fail/manual)

### 5.2 Monitor DynamoDB

```bash
aws dynamodb scan \
  --table-name vis_hosts \
  --region us-east-1

aws dynamodb scan \
  --table-name vis_cis_results \
  --region us-east-1 \
  --limit 10
```

### 5.3 Monitor Lambda Logs

```bash
aws logs tail /aws/lambda/visiblaze-ingest \
  --follow \
  --region us-east-1
```

### 5.4 Test API Endpoints Directly

```bash
# Get API URL from Terraform outputs
API_URL="https://your-api-gateway-url"
API_KEY="your-api-key"

# List hosts
curl -H "X-API-Key: ${API_KEY}" \
  "${API_URL}/hosts"

# Get specific host
curl -H "X-API-Key: ${API_KEY}" \
  "${API_URL}/hosts/{hostId}"

# Health check
curl "${API_URL}/health"
```

---

## Step 6: Scale and Monitoring

### 6.1 Launch Multiple Agents

Repeat Steps 4.1-4.5 for each EC2 instance. Each will send data to the same backend.

### 6.2 CloudWatch Monitoring

```bash
# View Lambda invocations
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Invocations \
  --dimensions Name=FunctionName,Value=visiblaze-ingest \
  --start-time 2025-11-11T00:00:00Z \
  --end-time 2025-11-11T23:59:59Z \
  --period 3600 \
  --statistics Sum \
  --region us-east-1

# View DynamoDB consumed write units
aws cloudwatch get-metric-statistics \
  --namespace AWS/DynamoDB \
  --metric-name ConsumedWriteCapacityUnits \
  --dimensions Name=TableName,Value=vis_hosts \
  --start-time 2025-11-11T00:00:00Z \
  --end-time 2025-11-11T23:59:59Z \
  --period 3600 \
  --statistics Sum
```

### 6.3 Set Up Alarms (Optional)

```bash
aws cloudwatch put-metric-alarm \
  --alarm-name visiblaze-lambda-errors \
  --alarm-description "Alert if Lambda has errors" \
  --metric-name Errors \
  --namespace AWS/Lambda \
  --statistic Sum \
  --period 300 \
  --threshold 5 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 1 \
  --dimensions Name=FunctionName,Value=visiblaze-ingest
```

---

## Step 7: Clean Up (Optional)

To remove all AWS resources:

```bash
cd infra/terraform

# Destroy (will delete all resources)
terraform destroy

# Confirm when prompted
```

---

## Troubleshooting

### Agent doesn't send data
1. Check config at `/etc/visiblaze-agent/config.yaml` (api_base_url, api_key)
2. Check agent logs: `tail -f /var/log/visiblaze-agent/agent.log`
3. Verify EC2 security group allows outbound HTTPS (port 443)
4. Test API manually: `curl -k https://your-api-gateway-url/health`

### Lambda returns 400/500
1. Check Lambda logs: `aws logs tail /aws/lambda/visiblaze-ingest --follow`
2. Verify DynamoDB tables exist: `aws dynamodb list-tables`
3. Check Lambda environment variables (API_KEY, etc.)
4. Test with sample payload: `curl -X POST ... -d '{"host":{...}}'`

### Frontend shows 404 on host detail
1. Ensure agent has been running and sent data (check DynamoDB)
2. Check API endpoint is correct (verify VITE_API_BASE_URL)
3. Open DevTools (F12) → Network tab to see actual API requests
4. Check CORS headers in API Gateway

### CloudFront shows 403
1. Ensure S3 bucket policy allows CloudFront access (Terraform should handle)
2. Invalidate CloudFront cache: `aws cloudfront create-invalidation --distribution-id ... --paths "/*"`
3. Wait 30 seconds for invalidation to complete

---

## Summary: Real-Time Workflow

1. **Hour 1**: Terraform applies (5-10 min), Lambda deploys (2-3 min), Frontend builds & uploads (3-5 min)
2. **Hour 1-2**: EC2 instance launches (2-3 min), agent installs (1-2 min), starts collecting (0 min first run)
3. **Real-Time Monitoring**: 
   - **Agent**: Logs visible in `/var/log/visiblaze-agent/agent.log` (refresh every interval)
   - **Backend**: Lambda logs in CloudWatch (refresh every 30 sec)
   - **Frontend**: Browser shows new hosts/checks as data arrives (refresh every 15 sec)
   - **Dashboard**: Real-time as agent sends, API processes, frontend fetches

Once deployed, visit the CloudFront URL and you'll see:
- Dashboard with live host data
- CIS compliance checks updating every 15 minutes (or interval set in config)
- Package inventory from all agents aggregated
- Compliance status and trends

---

## Next Steps

- Set up automated packaging in GitHub Actions (build & push to S3 on commit)
- Add custom CloudWatch dashboards for compliance trends
- Integrate with SNS for alerting on failed CIS checks
- Deploy multiple agent instances across regions for redundancy
