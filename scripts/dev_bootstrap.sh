#!/bin/bash
set -e

echo "=== Visiblaze Bootstrap Setup ==="
echo "Installing dependencies..."

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Installing Go 1.22..."
    wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
fi

# Check for Node
if ! command -v node &> /dev/null; then
    echo "Installing Node 18..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    sudo apt-get install -y nodejs
fi

# Check for Terraform
if ! command -v terraform &> /dev/null; then
    echo "Installing Terraform..."
    wget https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_linux_amd64.zip
    unzip terraform_1.5.0_linux_amd64.zip -d /usr/local/bin
fi

# Check for nfpm
if ! command -v nfpm &> /dev/null; then
    echo "Installing nfpm..."
    go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
fi

echo "✓ Dependencies installed"

echo "Building agent..."
make build-agent

echo "Building packages..."
make package

echo "Building Lambda function..."
cd backend/lambda
GOOS=linux GOARCH=amd64 go build -o bootstrap ./cmd/ingest
zip -r lambda_function.zip bootstrap
cd ../..

echo "Deploying infrastructure..."
make deploy-infra

# Extract outputs
API_ENDPOINT=$(terraform -chdir=infra/terraform output -raw api_endpoint)
API_KEY=$(terraform -chdir=infra/terraform output -raw api_key)

echo ""
echo "=== Setup Complete ==="
echo "API Endpoint: $API_ENDPOINT"
echo "API Key: $API_KEY"
echo ""
echo "Next steps:"
echo "  1. Update agent config: sudo nano /etc/visiblaze-agent/config.yaml"
echo "  2. Start frontend: cd web && npm run dev"
echo "  3. Tail agent logs: sudo journalctl -u visiblaze-agent -f"
