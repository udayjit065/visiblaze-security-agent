#!/bin/bash
# Visiblaze Automated Deployment Script
# Usage: ./deploy.sh [local|aws]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Step 1: Build Agent Binary
build_agent() {
    log_info "Building agent binary for Linux amd64..."
    
    mkdir -p "$REPO_ROOT/dist"
    
    cd "$REPO_ROOT/agent"
    
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags "-s -w -X main.Version=0.1.0" \
        -o ../dist/visiblaze-agent ./cmd/agent
    
    log_info "✓ Agent binary: $REPO_ROOT/dist/visiblaze-agent"
}

# Step 2: Build and Package
build_packages() {
    log_info "Building DEB package..."
    
    command -v nfpm >/dev/null || {
        log_warn "nfpm not found, installing..."
        go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
    }
    
    cd "$REPO_ROOT"
    nfpm package -f packaging/nfpm.yaml -p deb -t dist/ 2>/dev/null || \
        log_warn "DEB packaging failed (expected on Windows; use WSL or Linux)"
    
    ls -lh "$REPO_ROOT/dist/"* 2>/dev/null | grep -E "\.(deb|rpm|exe)" || true
}

# Step 3: Deploy with Terraform
terraform_deploy() {
    log_info "Preparing Terraform deployment..."
    
    if [ ! -f "$REPO_ROOT/infra/terraform/terraform.tfvars" ]; then
        log_warn "No terraform.tfvars found. Creating example..."
        cat > "$REPO_ROOT/infra/terraform/terraform.tfvars" << 'EOF'
aws_region              = "us-east-1"
project_name            = "visiblaze"
environment             = "prod"
lambda_memory           = 256
lambda_timeout          = 30
dynamodb_read_capacity  = 5
dynamodb_write_capacity = 5
enable_cloudfront       = true
EOF
        log_info "Created: $REPO_ROOT/infra/terraform/terraform.tfvars"
        log_warn "Edit terraform.tfvars before running terraform apply"
    fi
    
    cd "$REPO_ROOT/infra/terraform"
    
    log_info "Initializing Terraform..."
    terraform init
    
    log_info "Planning infrastructure changes..."
    terraform plan -out=tfplan
    
    echo ""
    read -p "Apply Terraform changes? (yes/no): " apply_confirm
    if [ "$apply_confirm" = "yes" ]; then
        terraform apply tfplan
        terraform output -json > "$REPO_ROOT/terraform_outputs.json"
        log_info "✓ Infrastructure deployed. Outputs saved to terraform_outputs.json"
    else
        log_warn "Terraform apply cancelled"
    fi
}

# Step 4: Build Lambda
build_lambda() {
    log_info "Building Lambda function..."
    
    cd "$REPO_ROOT/backend/lambda/cmd/ingest"
    
    GOOS=linux GOARCH=amd64 go build -o bootstrap .
    
    zip -q function.zip bootstrap
    
    log_info "✓ Lambda binary: function.zip"
}

# Step 5: Build and Upload Frontend
build_frontend() {
    log_info "Building React frontend..."
    
    cd "$REPO_ROOT/web"
    
    npm install --no-audit --no-fund >/dev/null 2>&1
    
    API_URL=$(grep -o '"api_url":"[^"]*' terraform_outputs.json 2>/dev/null | cut -d'"' -f4 || echo "http://localhost:3001")
    
    VITE_API_BASE_URL="$API_URL" npm run build 2>&1 | tail -20
    
    log_info "✓ Frontend built: $REPO_ROOT/web/dist"
    
    if [ -n "$1" ] && [ "$1" = "upload" ]; then
        log_info "Uploading to S3..."
        # TODO: aws s3 sync dist/ s3://bucket-name/ --delete
    fi
}

# Step 6: Test with mock server (local only)
test_local() {
    log_info "Starting mock ingest server..."
    
    cd "$REPO_ROOT/backend/mock"
    
    go run . &
    MOCK_PID=$!
    
    log_info "Mock server running (PID: $MOCK_PID)"
    
    sleep 2
    
    log_info "Testing agent..."
    cd "$REPO_ROOT"
    export VISIBLAZE_LOG_DIR="./logs"
    go run ./agent/cmd/agent -config ./agent/config.local.yaml -once
    
    log_info "✓ Agent test complete. Check logs/agent.log"
    
    log_info "Starting frontend dev server..."
    cd "$REPO_ROOT/web"
    export VITE_API_BASE_URL="http://localhost:3001"
    npm run dev &
    FRONTEND_PID=$!
    
    log_info "Frontend dev server running (PID: $FRONTEND_PID)"
    log_info "Open: http://localhost:5173"
    
    wait
}

# Main
case "${1:-local}" in
    local)
        log_info "Running LOCAL deployment (mock server + frontend dev)"
        build_agent
        test_local
        ;;
    aws)
        log_info "Running AWS deployment"
        build_agent
        build_packages
        build_lambda
        terraform_deploy
        build_frontend upload
        log_info "✓ AWS deployment complete"
        log_info "Frontend URL: $(grep cloudfront_domain_name terraform_outputs.json 2>/dev/null | head -1)"
        ;;
    *)
        log_error "Usage: $0 [local|aws]"
        exit 1
        ;;
esac
