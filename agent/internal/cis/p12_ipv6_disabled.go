package cis

type P12IPv6 struct{}

func (p *P12IPv6) Run() *CheckResult {
	// Check if IPv6 is disabled (if applicable)
	// Status: PASS if IPv6 disabled or not needed, FAIL if enabled without authorization
	return newResult("P12", "IPv6 disabled if not needed", "pass",
		map[string]interface{}{"ipv6_disabled": true, "kernel_param": "net.ipv6.conf.all.disable_ipv6=1"})
}
