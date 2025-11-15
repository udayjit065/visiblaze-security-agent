# Local Development Guide — Visiblaze Security Agent

Run the entire system locally without AWS for testing and development.

## Prerequisites

- Go 1.21+
- Node 18+ npm
- 3 terminal windows

## Step 1: Start Mock Ingest Server

The mock server simulates the AWS Lambda backend using file-based storage (no database needed).

```bash
cd d:/programming/dev/delta/projects/visiblaze-sec-agent/backend/mock
go run .
```

**Output**:
```
2025/11/11 20:21:31 mock server listening :3001 (data dir data)
```

The mock server now listens on `http://localhost:3001` and stores all payloads in `backend/mock/data/`.

## Step 2: Run Agent (One-Time Collection)

Open a new terminal and run the agent once to collect and send data:

```bash
cd d:/programming/dev/delta/projects/visiblaze-sec-agent
export VISIBLAZE_LOG_DIR=./logs
go run ./agent/cmd/agent -config ./agent/config.local.yaml -once
```

**What it does**:
1. Loads config from `agent/config.local.yaml` (points to http://localhost:3001)
2. Collects host info (hostname, OS, kernel, IP addresses)
3. Collects installed packages (dpkg, rpm, or apk depending on your OS)
4. Runs 13 CIS security compliance checks
5. POSTs JSON payload to `http://localhost:3001/ingest`
6. Logs everything to `./logs/agent.log`

**Check the result**:
```bash
# View agent logs
tail -f ./logs/agent.log

# View stored payload
cat backend/mock/data/*.json | jq .

# View mock server received it
# (should see "ingest" request in mock server terminal)
```

## Step 3: Start React Dashboard

Open a third terminal and start the frontend dev server:

```bash
cd d:/programming/dev/delta/projects/visiblaze-sec-agent/web
export VITE_API_BASE_URL=http://localhost:3001
npm install
npm run dev
```

**Output**:
```
  VITE v5.0.0  ready in 234 ms

  ➜  Local:   http://localhost:5173/
  ➜  press h to show help
```

## Step 4: View in Browser

Open your browser to `http://localhost:5173/`

You should see:
- **Hosts List**: Your machine listed with hostname, OS, kernel
- **Search Box**: Filter hosts by hostname, OS, or kernel version
- **Details Button**: Click to see:
  - CIS compliance check results (pass/fail/manual)
  - Package inventory
  - Host information

## How It Works Locally

```
┌──────────────┐
│ Agent        │  ← (Terminal 2) Collects data
│ (Go)         │
└──────┬───────┘
       │ HTTP POST to localhost:3001/ingest
       ▼
┌──────────────────────┐
│ Mock Server          │  ← (Terminal 1) Receives POST
│ (localhost:3001)     │    Stores in backend/mock/data/
└──────┬───────────────┘
       │
       │ GET /hosts, /hosts/{id}, /apps, /cis-results
       │
       ▼
┌──────────────────────┐
│ React Dashboard      │  ← (Terminal 3) Displays data
│ (localhost:5173)     │
└──────────────────────┘
```

## Development Workflow

### Run Agent Multiple Times

To send more data from agent:

```bash
# Terminal 2
export VISIBLAZE_LOG_DIR=./logs
go run ./agent/cmd/agent -config ./agent/config.local.yaml -once
```

Reload the frontend browser — it will fetch the new data.

### Inspect Payloads

All agent payloads are stored in `backend/mock/data/` as JSON files:

```bash
# List all payloads
ls backend/mock/data/

# View a payload
cat backend/mock/data/demo-host-1.json | jq .
# or
cat backend/mock/data/demo-host-1.json | jq '.host'
cat backend/mock/data/demo-host-1.json | jq '.cis_results'
cat backend/mock/data/demo-host-1.json | jq '.packages'
```

### Modify Agent Config

Edit `agent/config.local.yaml`:

```yaml
api_base_url: "http://localhost:3001"
api_key: "localtest"
collection_interval_minutes: 1
disable_ipv6_check: true
distro_hint: "linux"
```

Then re-run agent:
```bash
go run ./agent/cmd/agent -config ./agent/config.local.yaml -once
```

### Modify CIS Checks

Edit any check in `agent/internal/cis/p*.go`:

```go
// Example: make P1 always fail
func (p *P1PasswordQuality) Run() *CheckResult {
    return newResult("P1", "Password complexity", "fail", 
        map[string]interface{}{"reason": "testing"})
}
```

Rebuild and run agent:
```bash
go run ./agent/cmd/agent -config ./agent/config.local.yaml -once
```

### Modify Frontend

Edit React components in `web/src/components/`:
- `HostList.tsx` — host table with search
- `HostDetail.tsx` — host details view
- `CisResultsTable.tsx` — CIS checks display
- `PackagesTable.tsx` — package inventory

The Vite dev server auto-reloads when you save files.

## Common Tasks

### Test a Specific CIS Check

```bash
# Navigate to CIS check file
cd agent/internal/cis

# Edit the check (e.g., p5_firewall.go)
# Run agent
cd ../../../
go run ./agent/cmd/agent -config ./agent/config.local.yaml -once

# Check result
cat logs/agent.log | grep "P5"
```

### Test API Endpoints

```bash
# Mock server endpoints
curl http://localhost:3001/hosts | jq .
curl http://localhost:3001/health | jq .
curl http://localhost:3001/apps | jq .
curl http://localhost:3001/cis-results | jq .
```

### Run Agent Tests

```bash
cd agent
go test -v ./...
```

### Run Linter

```bash
# Install linter (one-time)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run it
cd agent
golangci-lint run ./...
```

### Build and Test Frontend Build

```bash
cd web

# Production build
npm run build

# Output is in web/dist/
# Ready to deploy to S3 + CloudFront (see AWS_DEPLOYMENT_GUIDE.md)
```

## Troubleshooting

### Mock Server Won't Start

```bash
# Check if port 3001 is already in use
netstat -an | grep 3001

# Kill existing process
pkill -f "go run"

# Start mock server again
cd backend/mock && go run .
```

### Agent Fails to Connect

```bash
# Check if mock server is running (should see output)
# Check agent config points to correct URL
cat agent/config.local.yaml

# Test connection manually
curl http://localhost:3001/health

# Check firewall
# (Windows Firewall might block; temporarily disable or add exception)
```

### Frontend Shows Empty

```bash
# Make sure mock server is running
# Make sure you ran agent with -once flag
# Check mock server data directory
ls backend/mock/data/

# If empty, agent didn't POST successfully
# Check agent logs
tail logs/agent.log
```

### TypeScript Errors in Frontend

```bash
# Reinstall dependencies
cd web
rm -rf node_modules
npm install

# Clear Vite cache
rm -rf .vite

# Restart dev server
npm run dev
```

## File Structure (Local Data)

```
project-root/
  logs/                           # Agent logs
    agent.log                     # JSON-formatted logs
  backend/mock/
    data/                         # Stored payloads from agents
      demo-host-1.json           # Example payload
      my-computer.json           # Your actual machine data (after running agent)
    main.go                       # Mock server source
```

## Performance Notes

- Mock server is memory-resident (no database overhead)
- Storing payloads as files is **slow at scale** (fine for <100 hosts locally)
- For AWS deployment with thousands of hosts, use Lambda + DynamoDB (see AWS_DEPLOYMENT_GUIDE.md)

## Moving to AWS

When you're ready to deploy to AWS:

1. Follow [AWS_DEPLOYMENT_GUIDE.md](./AWS_DEPLOYMENT_GUIDE.md)
2. Or run: `./deploy.sh aws`

The same agent code works unchanged; just change config to point to Lambda API URL instead of localhost.

## Integration Testing

Test the entire flow locally:

```bash
#!/bin/bash
# save as test_local.sh

# 1. Start mock server
cd backend/mock && go run . &
MOCK_PID=$!
sleep 1

# 2. Run agent
export VISIBLAZE_LOG_DIR=./logs
go run ./agent/cmd/agent -config ./agent/config.local.yaml -once

# 3. Verify payload exists
if [ -f "backend/mock/data/"*.json ]; then
  echo "✓ Payload stored"
else
  echo "✗ Payload not found"
fi

# 4. Verify mock server endpoints
curl -s http://localhost:3001/hosts | jq . > /dev/null && echo "✓ /hosts works" || echo "✗ /hosts failed"
curl -s http://localhost:3001/health | jq . > /dev/null && echo "✓ /health works" || echo "✗ /health failed"

# Cleanup
kill $MOCK_PID

echo "✓ Local integration test complete"
```

Run it:
```bash
bash test_local.sh
```

## Next Steps

- ✅ **Understand flow locally** — You've done this by running the system
- ➜ **Deploy to AWS** — See [AWS_DEPLOYMENT_GUIDE.md](./AWS_DEPLOYMENT_GUIDE.md)
- ➜ **Add custom CIS checks** — Edit `agent/internal/cis/p*.go`
- ➜ **Customize dashboard** — Edit `web/src/components/`
- ➜ **Monitor in production** — Set up CloudWatch alerts and dashboards

## Support

- Check agent logs: `tail -f logs/agent.log`
- Check mock server output in terminal 1
- Check frontend console: Press F12 in browser
- Look for 404 errors in Network tab of DevTools
