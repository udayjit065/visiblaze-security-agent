package collect

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type HostInfo struct {
	HostID       string   `json:"host_id"`
	Hostname     string   `json:"hostname"`
	OSID         string   `json:"os_id"`
	OSVersion    string   `json:"os_version"`
	Kernel       string   `json:"kernel"`
	IPAddresses  []string `json:"ip_addresses"`
	AgentVersion string   `json:"agent_version"`
}

func GetHostInfo(agentVersion string) (*HostInfo, error) {
	hostID, err := getOrCreateHostID()
	if err != nil {
		return nil, fmt.Errorf("get host id: %w", err)
	}

	hostname, _ := os.Hostname()
	osID, osVersion := detectOS()
	kernel, _ := util.RunCmd("uname", "-r")
	ips := getIPAddresses()

	return &HostInfo{
		HostID:       hostID,
		Hostname:     hostname,
		OSID:         osID,
		OSVersion:    osVersion,
		Kernel:       strings.TrimSpace(kernel),
		IPAddresses:  ips,
		AgentVersion: agentVersion,
	}, nil
}

func getOrCreateHostID() (string, error) {
	idDir := "/var/lib/visiblaze-agent"
	idFile := fmt.Sprintf("%s/host_id", idDir)

	if content, err := util.ReadFile(idFile); err == nil && strings.TrimSpace(content) != "" {
		return strings.TrimSpace(content), nil
	}

	newID := uuid.New().String()
	if err := util.EnsureDir(idDir); err != nil {
		return "", err
	}
	if err := util.EnsureFile(idFile, newID); err != nil {
		return "", err
	}

	return newID, nil
}

func detectOS() (string, string) {
	osRelease := "/etc/os-release"
	lines, err := util.ReadFileLines(osRelease)
	if err != nil {
		return "unknown", "unknown"
	}

	osID := "unknown"
	osVersion := "unknown"

	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			osID = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			osVersion = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), `"`)
		}
	}

	return osID, osVersion
}

func getIPAddresses() []string {
	output, err := util.RunCmd("ip", "addr", "show")
	if err != nil {
		return []string{}
	}

	var ips []string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "inet ") && !strings.Contains(line, "127.0.0.1") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ipWithCIDR := parts[1]
				ip := strings.Split(ipWithCIDR, "/")[0]
				ips = append(ips, ip)
			}
		}
	}

	return ips
}
