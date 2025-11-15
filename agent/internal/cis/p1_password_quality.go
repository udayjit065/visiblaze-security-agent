package cis

import (
	"strings"

	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type P1PasswordQuality struct{}

func (p *P1PasswordQuality) Run() *CheckResult {
	pwqualityConf := "/etc/security/pwquality.conf"
	content, err := util.ReadFile(pwqualityConf)
	if err != nil {
		return newResult("P1", "Password complexity enforced", "manual",
			map[string]interface{}{"reason": "pwquality.conf not found"})
	}

	evidence := make(map[string]interface{})
	minlen, dcredit, ucredit, lcredit, ocredit := 0, 0, 0, 0, 0

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if strings.Contains(line, "minlen") {
			evidence["minlen"] = line
			if strings.Contains(line, "14") || strings.Contains(line, "15") {
				minlen = 1
			}
		}
		if strings.Contains(line, "dcredit") {
			evidence["dcredit"] = line
			if strings.Contains(line, "-") {
				dcredit = 1
			}
		}
		if strings.Contains(line, "ucredit") {
			evidence["ucredit"] = line
			if strings.Contains(line, "-") {
				ucredit = 1
			}
		}
		if strings.Contains(line, "lcredit") {
			evidence["lcredit"] = line
			if strings.Contains(line, "-") {
				lcredit = 1
			}
		}
		if strings.Contains(line, "ocredit") {
			evidence["ocredit"] = line
			if strings.Contains(line, "-") {
				ocredit = 1
			}
		}
	}

	if minlen+dcredit+ucredit+lcredit+ocredit >= 4 {
		return newResult("P1", "Password complexity enforced", "pass", evidence)
	}
	return newResult("P1", "Password complexity enforced", "fail", evidence)
}
