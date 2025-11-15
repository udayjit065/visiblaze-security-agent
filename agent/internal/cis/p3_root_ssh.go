package cis

import (
	"strings"

	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type P3RootSSH struct{}

func (p *P3RootSSH) Run() *CheckResult {
	sshConfig := "/etc/ssh/sshd_config"
	content, err := util.ReadFile(sshConfig)
	if err != nil {
		return newResult("P3", "Root login over SSH disabled", "manual",
			map[string]interface{}{"reason": "sshd_config not found"})
	}

	evidence := make(map[string]interface{})
	permitRootLine := ""

	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "PermitRootLogin") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			permitRootLine = line
			evidence["PermitRootLogin"] = line
			break
		}
	}

	if strings.Contains(permitRootLine, "no") {
		return newResult("P3", "Root login over SSH disabled", "pass", evidence)
	}
	return newResult("P3", "Root login over SSH disabled", "fail", evidence)
}
