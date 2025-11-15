package cis

import (
	"time"
)

type CheckResult struct {
	CheckID   string                 `json:"check_id"`
	Title     string                 `json:"title"`
	Status    string                 `json:"status"`
	Evidence  map[string]interface{} `json:"evidence"`
	Timestamp string                 `json:"ts"`
}

type CheckRunner interface {
	Run() *CheckResult
}

func RunAllChecks() []*CheckResult {
	runners := []CheckRunner{
		&P1PasswordQuality{},
		&P2PasswordExpiry{},
		&P3RootSSH{},
		&P4UnusedFS{},
		&P5Firewall{},
		&P6TimeSync{},
		&P7Auditd{},
		&P8MAC{},
		&P9WorldWritable{},
		&P10GDMAutoLogin{},
		&P11SSHProtocol2{},
		&P12IPv6{},
	}

	results := make([]*CheckResult, 0, len(runners))
	for _, runner := range runners {
		result := runner.Run()
		result.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
		results = append(results, result)
	}

	return results
}

func newResult(checkID, title, status string, evidence map[string]interface{}) *CheckResult {
	if evidence == nil {
		evidence = make(map[string]interface{})
	}
	return &CheckResult{
		CheckID:  checkID,
		Title:    title,
		Status:   status,
		Evidence: evidence,
	}
}
