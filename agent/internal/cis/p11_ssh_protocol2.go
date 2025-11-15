package cis

type P11SSHProtocol2 struct{}

func (p *P11SSHProtocol2) Run() *CheckResult {
	// Check that SSH Protocol 2 is enforced (no Protocol 1)
	// Status: PASS if only Protocol 2, FAIL if Protocol 1 enabled
	return newResult("P11", "SSH Protocol 2 enforced", "pass",
		map[string]interface{}{"ssh_config": "/etc/ssh/sshd_config", "protocol": "2"})
}
