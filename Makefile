.PHONY: help build-agent test lint package deploy-infra run-agent web clean

AGENT_BINARY := visiblaze-agent
AGENT_VERSION := 0.1.0
GOFLAGS := -ldflags "-s -w -X main.Version=$(AGENT_VERSION)"

help:
	@echo "Visiblaze Build Targets"
	@echo "  make build-agent          Build agent binary"
	@echo "  make test                 Run tests + lint"
	@echo "  make lint                 Run linter"
	@echo "  make package              Build deb/rpm"
	@echo "  make deploy-infra         Deploy Terraform"
	@echo "  make run-agent            Run agent locally"
	@echo "  make web                  React dev server"
	@echo "  make clean                Remove artifacts"

build-agent:
	@echo "Building agent..."
	cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o ../dist/$(AGENT_BINARY) ./cmd/agent
	@echo "✓ Binary: dist/$(AGENT_BINARY)"

test:
	@echo "Running tests..."
	cd agent && go test -v -race -coverprofile=coverage.out ./...

lint:
	@command -v golangci-lint >/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./agent/...

package: build-agent
	@echo "Building packages..."
	@command -v nfpm >/dev/null || go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
	mkdir -p dist
	nfpm package -f packaging/nfpm.yaml -p deb -o dist/
	nfpm package -f packaging/nfpm.yaml -p rpm -o dist/
	@ls -lh dist/*.{deb,rpm} 2>/dev/null

deploy-infra:
	@echo "Deploying Terraform..."
	cd infra/terraform && terraform init && terraform apply -auto-approve
	cd infra/terraform && terraform output -json > /tmp/visiblaze_outputs.json
	@echo "✓ Outputs saved to /tmp/visiblaze_outputs.json"

run-agent: build-agent
	@test -z "$$VISIBLAZE_CONFIG" && echo "ERROR: Set VISIBLAZE_CONFIG" && exit 1
	./dist/$(AGENT_BINARY) -config $$VISIBLAZE_CONFIG -once

web:
	cd web && npm install && npm run dev

clean:
	rm -rf dist/ agent/coverage.out
	cd agent && go clean
