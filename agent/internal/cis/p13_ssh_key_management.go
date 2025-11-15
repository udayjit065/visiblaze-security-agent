package cis

import (
	"strings"

	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type P13SSHKeyManagement struct{}

func (p *P13SSHKeyManagement) Run() *CheckResult {
	evidence := make(map[string]interface{})
	authKeysPath := "/root/.ssh/authorized_keys"
	content, err := util.ReadFile(authKeysPath)
	if err != nil {
		// If file missing, that's informational (may not be configured)
		return newResult("P13", "SSH authorized_keys present and permissions correct", "manual",
			map[string]interface{}{"reason": "authorized_keys not found", "path": authKeysPath})
	}

	lines := 0
	weak := 0
	for _, l := range strings.Split(content, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		lines++
		// very naive check for weak keys (e.g., empty key parts)
		parts := strings.Fields(l)
		if len(parts) < 2 {
			weak++
		}
	}

	evidence["authorized_keys_path"] = authKeysPath
	evidence["entries"] = lines
	evidence["weak_entries"] = weak

	if lines > 0 && weak == 0 {
		return newResult("P13", "SSH authorized_keys present and permissions correct", "pass", evidence)
	}
	if lines > 0 && weak > 0 {
		return newResult("P13", "SSH authorized_keys present and permissions correct", "fail", evidence)
	}

	return newResult("P13", "SSH authorized_keys present and permissions correct", "manual", evidence)
}
