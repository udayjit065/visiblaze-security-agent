package cis

import (
	"strings"

	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type P6TimeSync struct{}

func (p *P6TimeSync) Run() *CheckResult {
	evidence := make(map[string]interface{})

	// Check chrony
	chronyStatus, _ := util.RunCmd("systemctl", "is-active", "chronyd")
	evidence["chronyd"] = chronyStatus
	if strings.Contains(chronyStatus, "active") {
		return newResult("P6", "Time sync configured", "pass", evidence)
	}

	// Check ntpd
	ntpdStatus, _ := util.RunCmd("systemctl", "is-active", "ntpd")
	evidence["ntpd"] = ntpdStatus
	if strings.Contains(ntpdStatus, "active") {
		return newResult("P6", "Time sync configured", "pass", evidence)
	}

	return newResult("P6", "Time sync configured", "fail", evidence)
}
