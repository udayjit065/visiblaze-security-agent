# Visiblaze Security Agent

Lightweight security compliance agent that collects host information, installed packages, and runs CIS Level 1 checks. Data is ingested into a backend (AWS Lambda or mock server) and displayed via a React frontend dashboard.

## Quick Links    

- **[AWS Deployment Guide](./AWS_DEPLOYMENT_GUIDE.md)** — Deploy to AWS (Lambda, DynamoDB, CloudFront)
- **[Local Development Guide](./LOCAL_DEV.md)** — Run locally with mock server
- **[Deploy Script](./deploy.sh)** — Automated one-command deployment

## Quick Start: Local (No AWS Needed)

Run everything locally with mock server in 3 terminals:

```bash
# Terminal 1: Start mock ingest server
cd backend/mock && go run .

# Terminal 2: Run agent (collects data)
export VISIBLAZE_LOG_DIR=./logs
go run ./agent/cmd/agent -config ./agent/config.local.yaml -once

# Terminal 3: Start React dashboard
cd web
export VITE_API_BASE_URL=http://localhost:3001
npm install && npm run dev

# Open browser: http://localhost:5173
```

## Quick Start: AWS Deployment

Deploy everything to AWS:

```bash
# Prerequisites: AWS CLI configured, Terraform installed, Go 1.21+, Node 18+

cd d:/programming/dev/delta/projects/visiblaze-sec-agent

# Automated: builds agent, deploys infrastructure, launches frontend
./deploy.sh aws

# Manual backend deploy helper commands
./scripts/build_lambda.sh
terraform -chdir=infra/terraform init
terraform -chdir=infra/terraform apply

# Or follow the detailed guide: AWS_DEPLOYMENT_GUIDE.md
```

**Real-time monitoring**: Once deployed, visit the CloudFront URL to see:
- Live host list with compliance status
- CIS check results updating every 15 minutes
- Package inventory aggregated across all agents
- Real-time logs in CloudWatch

## Architecture

```
Agent (EC2 / Linux)
  ├─ Collects: Host info, packages, CIS compliance checks
  └─ Sends every 15 min to Lambda API via HTTPS + API Key
                        ↓
            Lambda API Gateway  
                        ↓
    DynamoDB Tables (vis_hosts, vis_packages, vis_cis_results)
                        ↓
            CloudFront CDN + S3
                        ↓
        React Dashboard (Browser)
                        
   All data visible in real-time as agents report
```

## Project Structure

```
agent/                     # Security compliance agent (Go)
  cmd/agent/main.go        # CLI entry point
  internal/
    cis/                   # 13 CIS Level 1 compliance checks
    collect/               # Host info & package collector
    config/                # YAML config loader
    ingest/                # API client
    logging/               # JSON structured logging
    schedule/              # Scheduled collection (15 min intervals)

backend/
  lambda/                  # AWS Lambda handlers (Go)
    cmd/ingest/main.go     # Lambda entry point (deployed to AWS)
    internal/handlers/     # /ingest, /hosts, /apps, /cis-results, /health
  mock/main.go             # Local test server (no AWS, file-based storage)

infra/terraform/           # AWS infrastructure as code
  api_gateway.tf           # REST API + routes
  dynamodb.tf              # Three tables
  lambda.tf                # Function + execution role
  iam.tf                   # Permissions
  main.tf                  # Core resources
  variables.tf             # Configurable parameters

web/                       # React frontend (TypeScript + Vite)
  src/
    components/
      HostList.tsx         # List all hosts (with search filter)
      HostDetail.tsx       # Host details + CIS results + packages
      CisResultsTable.tsx  # Compliance check table
      PackagesTable.tsx    # Package inventory table
    lib/api.ts             # Axios client
    App.tsx                # Main app component

packaging/                 # DEB/RPM package config
  nfpm.yaml                # Package metadata & installation scripts

scripts/                   # Build helpers
  build_all.sh
  dev_bootstrap.sh
  install_local_deb.sh
```

## 13 CIS Level 1 Security Checks

Each check runs on agent and reports pass/fail/manual status:

| # | Name | Category |
|---|------|----------|
| P1 | Password complexity enforced | Account & Access |
| P2 | Password expiration policy | Account & Access |
| P3 | Root SSH login disabled | Account & Access |
| P4 | Unused filesystems disabled | System Hardening |
| P5 | Firewall enabled | Network Security |
| P6 | Time sync configured | System Config |
| P7 | Auditd installed & enabled | Logging |
| P8 | Mandatory Access Control enforced | System Hardening |
| P9 | No world-writable files | File Permissions |
| P10 | GDM autologin disabled | Account & Access |
| P11 | SSH Protocol 2 enforced | Network Security |
| P12 | IPv6 disabled if not needed | Network Security |
| P13 | SSH authorized_keys properly managed | Account & Access |

## Real-Time Flow Example

1. **Agent starts** (runs every 15 minutes or on-demand with `-once`)
   - Collects host info (hostname, OS, kernel, IP addresses)
   - Gathers installed packages (dpkg, rpm, apk)
   - Executes 13 CIS compliance checks
   - POSTs JSON payload to Lambda API with X-API-Key header

2. **Lambda receives** ingest request
   - Validates API key
   - Parses JSON payload
   - Stores in DynamoDB (atomic write to 3 tables)
   - Returns 200 OK

3. **Frontend fetches** data (on page load or auto-refresh)
   - GET /hosts → lists all monitored systems
   - GET /hosts/{hostId} → shows single host with CIS results & packages
   - GET /apps → package inventory
   - GET /cis-results → compliance dashboard
   - All API calls hit CloudFront cache or Lambda

4. **Dashboard displays**
   - Host list with last-seen timestamp, OS, kernel
   - Click "Details" → shows CIS check status, evidence
   - Search/filter hosts by hostname, OS, kernel
   - All data refreshes as agents report

**Result**: Compliance visibility across entire infrastructure in real-time.

## Supported Distributions

- Ubuntu 20.04+ (uses dpkg for package detection)
- Debian 10+ (uses dpkg)
- RHEL 7+ (uses rpm)
- CentOS 7+ (uses rpm)
- Alpine (uses apk)
- Amazon Linux 2 (uses rpm)

## Technologies

| Component | Stack |
|-----------|-------|
| Agent | Go 1.21+, systemd |
| Backend | AWS Lambda (Go), DynamoDB, API Gateway |
| Frontend | React 18, TypeScript, Vite, Axios |
| Infrastructure | Terraform, CloudFront, S3, IAM |
| Packaging | nfpm (DEB/RPM), systemd unit |
| Logging | CloudWatch (Lambda), JSON (Agent) |

## Key Commands

```bash
# Build agent binary for Linux
cd agent && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../dist/visiblaze-agent ./cmd/agent

# Create DEB package
nfpm package -f packaging/nfpm.yaml -p deb -t dist/

# Test agent locally
export VISIBLAZE_LOG_DIR=./logs
go run ./agent/cmd/agent -config ./agent/config.local.yaml -once

# Deploy infrastructure
cd infra/terraform && terraform init && terraform apply

# Build frontend
cd web && npm install && npm run build

# Start mock server (for testing without AWS)
cd backend/mock && go run .

# View Lambda logs
aws logs tail /aws/lambda/visiblaze-ingest --follow

# Query DynamoDB
aws dynamodb scan --table-name vis_hosts --region us-east-1
```

## Configuration

### Agent Config (`agent/config.yaml` on server)
```yaml
api_base_url: "https://your-lambda-api-url"
api_key: "your-secure-api-key"
collection_interval_minutes: 15
disable_ipv6_check: false
distro_hint: "ubuntu"
```

### Local Dev Config (`agent/config.local.yaml`)
```yaml
api_base_url: "http://localhost:3001"
api_key: "localtest"
collection_interval_minutes: 1
disable_ipv6_check: true
distro_hint: "linux"
```

## Deployment Timeline

- **5-10 min**: Terraform provisioning (Lambda, DynamoDB, API Gateway, CloudFront)
- **2-3 min**: Frontend build & upload to S3
- **1-2 min**: EC2 instance launch & agent install
- **0 min**: Agent starts collecting (first run immediate)
- **Real-time**: Data flows through pipeline as agent sends reports

After deployment:
- **Every 15 min**: Agent collects & sends new data
- **Live**: Dashboard updates as data arrives
- **Searchable**: Filter hosts by name, OS, kernel
- **Monitoring**: CloudWatch dashboards for Lambda & DynamoDB metrics

## Troubleshooting

**Agent doesn't send data**
- Check config: `cat /etc/visiblaze-agent/config.yaml`
- Check logs: `tail -f /var/log/visiblaze-agent/agent.log`
- Verify network: `curl -k https://your-api-url/health`

**Frontend shows 404**
- Ensure agent has sent data (check DynamoDB: `aws dynamodb scan --table-name vis_hosts`)
- Verify API base URL in frontend config
- Check CloudFront is caching correctly

**Lambda returns error**
- Check logs: `aws logs tail /aws/lambda/visiblaze-ingest --follow`
- Verify DynamoDB tables exist: `aws dynamodb list-tables`
- Test API: `curl -H "X-API-Key: yourkey" https://your-api/health`

See **[AWS_DEPLOYMENT_GUIDE.md](./AWS_DEPLOYMENT_GUIDE.md)** for detailed troubleshooting.

## Features

✅ **13 CIS Level 1 Compliance Checks** — Industry-standard security baselines  
✅ **Multi-OS Support** — Ubuntu, Debian, RHEL, CentOS, Alpine, Amazon Linux  
✅ **Real-Time Dashboard** — Live compliance status across all hosts  
✅ **Package Inventory** — Track installed packages across infrastructure  
✅ **Scalable Backend** — Serverless Lambda handles thousands of agents  
✅ **Secure APIs** — API Key authentication, HTTPS only  
✅ **Infrastructure as Code** — Reproducible Terraform deployments  
✅ **Search & Filter** — Find hosts by name, OS, kernel version  
✅ **CloudFront CDN** — Global distribution of dashboard  
✅ **JSON Logging** — Structured logs for analysis  

## Next Steps

1. **Deploy Locally**: Follow Quick Start → Local to test everything
2. **Deploy to AWS**: Follow AWS Deployment Guide or run `./deploy.sh aws`
3. **Monitor**: View real-time data on dashboard as agents report
4. **Scale**: Launch agents on multiple EC2 instances for organization-wide visibility
5. **Customize**: Modify CIS checks or add new ones in `agent/internal/cis/`

## License

MIT

## Support

- Issues? Check [AWS_DEPLOYMENT_GUIDE.md](./AWS_DEPLOYMENT_GUIDE.md) troubleshooting
- Local dev? See [LOCAL_DEV.md](./LOCAL_DEV.md)
- Questions? Review agent logs or Lambda CloudWatch logs
