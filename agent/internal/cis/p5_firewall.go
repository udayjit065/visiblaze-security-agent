package cis

import (
	"strings"

	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type P5Firewall struct{}

func (p *P5Firewall) Run() *CheckResult {
	evidence := make(map[string]interface{})

	// Try UFW (Ubuntu)
	ufwStatus, _ := util.RunCmd("ufw", "status")
	evidence["ufw_status"] = ufwStatus

	if strings.Contains(ufwStatus, "active") {
		return newResult("P5", "Firewall enabled", "pass", evidence)
	}

	// Try firewalld (RHEL)
	fwStatus, _ := util.RunCmd("systemctl", "is-active", "firewalld")
	evidence["firewalld_status"] = fwStatus

	if strings.Contains(fwStatus, "active") {
		return newResult("P5", "Firewall enabled", "pass", evidence)
	}

	return newResult("P5", "Firewall enabled", "fail", evidence)
}
