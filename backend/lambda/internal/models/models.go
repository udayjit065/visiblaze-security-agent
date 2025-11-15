package models

type Host struct {
	HostID       string   `json:"host_id"`
	Hostname     string   `json:"hostname"`
	OSID         string   `json:"os_id"`
	OSVersion    string   `json:"os_version"`
	Kernel       string   `json:"kernel"`
	IPAddresses  []string `json:"ip_addresses"`
	AgentVersion string   `json:"agent_version"`
}

type Package struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Arch        string `json:"arch"`
	Manager     string `json:"manager"`
	Source      string `json:"source"`
	InstalledAt string `json:"installed_at"`
}

type CISResult struct {
	CheckID   string                 `json:"check_id"`
	Title     string                 `json:"title"`
	Status    string                 `json:"status"`
	Evidence  map[string]interface{} `json:"evidence"`
	Timestamp string                 `json:"ts"`
}

type IngestPayload struct {
	Host       Host        `json:"host"`
	Packages   []Package   `json:"packages"`
	CISResults []CISResult `json:"cis_results"`
}
