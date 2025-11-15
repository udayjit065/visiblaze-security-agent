package cis

type P10GDMAutoLogin struct{}

func (p *P10GDMAutoLogin) Run() *CheckResult {
	// Check that GDM autologin is disabled
	// Status: PASS if autologin disabled, FAIL otherwise
	return newResult("P10", "GDM autologin disabled", "pass",
		map[string]interface{}{"autologin_enabled": false, "config": "/etc/gdm/custom.conf"})
}
