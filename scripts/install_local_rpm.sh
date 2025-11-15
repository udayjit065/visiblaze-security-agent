#!/bin/bash

if [ ! -f "dist/visiblaze-agent-"*.rpm ]; then
    echo "Error: No .rpm found in dist/. Run 'make package' first"
    exit 1
fi

RPM=$(ls dist/visiblaze-agent-*.rpm | head -1)
echo "Installing $RPM..."

sudo rpm -ivh "$RPM"

echo "✓ Installed"
echo ""
echo "Configuration: /etc/visiblaze-agent/config.yaml"
echo "Logs: journalctl -u visiblaze-agent -f"
sudo systemctl status visiblaze-agent
