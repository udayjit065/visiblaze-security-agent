#!/bin/bash

if [ ! -f "dist/visiblaze-agent_"*.deb ]; then
    echo "Error: No .deb found in dist/. Run 'make package' first"
    exit 1
fi

DEB=$(ls dist/visiblaze-agent_*.deb | head -1)
echo "Installing $DEB..."

sudo dpkg -i "$DEB" || sudo apt-get install -f -y

echo "✓ Installed"
echo ""
echo "Configuration: /etc/visiblaze-agent/config.yaml"
echo "Logs: journalctl -u visiblaze-agent -f"
sudo systemctl status visiblaze-agent
