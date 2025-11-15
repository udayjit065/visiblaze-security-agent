package cis

import (
	"strings"

	"github.com/visiblaze/sec-agent/agent/internal/util"
)

type P4UnusedFS struct{}

func (p *P4UnusedFS) Run() *CheckResult {
	evidence := make(map[string]interface{})
	fsToCheck := []string{"cramfs", "squashfs", "udf"}

	procFS, _ := util.ReadFile("/proc/filesystems")
	evidence["supported_filesystems"] = procFS

	blacklistedCount := 0
	for _, fs := range fsToCheck {
		if !strings.Contains(procFS, fs) {
			blacklistedCount++
		}
	}

	evidence["blacklisted_count"] = blacklistedCount

	if blacklistedCount == len(fsToCheck) {
		return newResult("P4", "Unused filesystems disabled", "pass", evidence)
	}
	return newResult("P4", "Unused filesystems disabled", "fail", evidence)
}
