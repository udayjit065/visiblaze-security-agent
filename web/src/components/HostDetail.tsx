import { useEffect, useState } from 'react'
import { fetchHostDetail } from '../lib/api'
import { Host } from '../types'
import CisResultsTable from './CisResultsTable'
import './HostDetail.css'
import PackagesTable from './PackagesTable'

interface Props {
  hostId: string
  onBack: () => void
}

export default function HostDetail({ hostId, onBack }: Props) {
  const [host, setHost] = useState<Host | null>(null)
  const [tab, setTab] = useState<'packages' | 'cis'>('packages')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    fetchHostDetail(hostId)
      .then(res => {
        // backend/mock returns host object directly
        const data = res.data as any
        if (data.host) setHost(data.host as Host)
        else setHost(data as Host)
        setLoading(false)
      })
      .catch(err => {
        setError(err.message || 'failed to load host')
        setLoading(false)
      })
  }, [hostId])

  return (
    <div className="host-detail">
      <button onClick={onBack} className="btn btn-secondary">
        ← Back
      </button>
      
  {loading && <div>Loading host...</div>}
  {error && <div className="error">Error: {error}</div>}

  {host && (
        <>
          <h2>{host.hostname}</h2>
          <div className="host-info">
            <p><strong>OS:</strong> {host.os_id} {host.os_version}</p>
            <p><strong>Kernel:</strong> {host.kernel}</p>
            <p><strong>Agent:</strong> {host.agent_version}</p>
            <p><strong>IPs:</strong> {host.ip_addresses && host.ip_addresses.length > 0 ? host.ip_addresses.join(', ') : '-'}</p>
          </div>

          <div className="tabs">
            <button
              className={`tab-btn ${tab === 'packages' ? 'active' : ''}`}
              onClick={() => setTab('packages')}
            >
              Packages
            </button>
            <button
              className={`tab-btn ${tab === 'cis' ? 'active' : ''}`}
              onClick={() => setTab('cis')}
            >
              CIS Checks
            </button>
          </div>

          {tab === 'packages' && <PackagesTable hostId={hostId} />}
          {tab === 'cis' && <CisResultsTable hostId={hostId} />}
        </>
      )}
    </div>
  )
}
