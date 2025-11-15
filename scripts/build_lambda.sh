#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LAMBDA_DIR="${ROOT_DIR}/backend/lambda"
BUILD_DIR="./backend/lambda/cmd/ingest" # <-- This line is modified
OUTPUT="${LAMBDA_DIR}/bootstrap"

echo "→ Building Lambda binary (target: ${OUTPUT})"

# --- THIS IS THE FIX ---
# Change directory to the root of the Go module so build paths are correct
cd "${ROOT_DIR}" 
# --- END FIX ---

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "${OUTPUT}" "${BUILD_DIR}"
chmod +x "${OUTPUT}"

echo "✓ Lambda bootstrap built."
echo
echo "Next steps:"
echo "  1. Run 'terraform -chdir=${ROOT_DIR}/infra/terraform init'"
echo "  2. Run 'terraform -chdir=${ROOT_DIR}/infra/terraform apply'"
echo "Terraform will package ${OUTPUT} into lambda_function.zip automatically."