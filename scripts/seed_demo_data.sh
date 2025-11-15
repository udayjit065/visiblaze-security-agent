#!/bin/bash

# Seeds demo host data to the API

API_ENDPOINT="${1:?Usage: $0 <API_ENDPOINT> [API_KEY]}"
API_KEY="${2:?Usage: $0 <API_ENDPOINT> <API_KEY>}"

HOST_ID="demo-$(uuidgen)"

PAYLOAD=$(cat <<EOF
{
  "host": {
    "host_id": "$HOST_ID",
    "hostname": "demo-host",
    "os_id": "ubuntu",
    "os_version": "22.04",
    "kernel": "5.15.0-generic",
    "ip_addresses": ["192.168.1.100"],
    "agent_version": "0.1.0"
  },
  "packages": [
    {"name":"curl","version":"7.81.0","arch":"amd64","manager":"dpkg","source":"ubuntu"},
    {"name":"openssh-server","version":"1:8.2p1","arch":"amd64","manager":"dpkg","source":"ubuntu"}
  ],
  "cis_results": [
    {"check_id":"P1","title":"Password complexity","status":"pass","evidence":{"minlen":"14"},"ts":"2024-01-01T00:00:00Z"},
    {"check_id":"P3","title":"Root SSH disabled","status":"pass","evidence":{"PermitRootLogin":"no"},"ts":"2024-01-01T00:00:00Z"}
  ]
}
EOF
)

echo "Sending to $API_ENDPOINT..."
curl -X POST \
  "$API_ENDPOINT/ingest" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "$PAYLOAD"

echo ""
echo "✓ Demo data sent"
