export interface Host {
  host_id: string
  hostname: string
  os_id: string
  os_version: string
  kernel: string
  ip_addresses: string[]
  agent_version: string
  last_seen: string
  first_seen: string
}

export interface Package {
  name: string
  version: string
  arch: string
  manager: string
  source: string
  installed_at?: string
}

export interface CISResult {
  check_id: string
  title: string
  status: 'pass' | 'fail' | 'manual'
  evidence: Record<string, any>
  ts: string
}

export interface IngestPayload {
  host: Host
  packages: Package[]
  cis_results: CISResult[]
}
