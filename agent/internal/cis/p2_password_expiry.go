package cis

import (
	"strings"

	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type P2PasswordExpiry struct{}

func (p *P2PasswordExpiry) Run() *CheckResult {
	loginDefs := "/etc/login.defs"
	content, err := util.ReadFile(loginDefs)
	if err != nil {
		return newResult("P2", "Password expiration policy", "manual",
			map[string]interface{}{"reason": "login.defs not found"})
	}

	evidence := make(map[string]interface{})
	passed := 0

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "PASS_MAX_DAYS") {
			evidence["PASS_MAX_DAYS"] = line
			if !strings.Contains(line, "99999") {
				passed++
			}
		}
		if strings.HasPrefix(strings.TrimSpace(line), "PASS_MIN_DAYS") {
			evidence["PASS_MIN_DAYS"] = line
			if !strings.Contains(line, "0") {
				passed++
			}
		}
		if strings.HasPrefix(strings.TrimSpace(line), "PASS_WARN_AGE") {
			evidence["PASS_WARN_AGE"] = line
			if strings.Contains(line, "7") {
				passed++
			}
		}
	}

	if passed >= 3 {
		return newResult("P2", "Password expiration policy", "pass", evidence)
	}
	return newResult("P2", "Password expiration policy", "fail", evidence)
}
