import { useEffect, useState } from 'react'
import { fetchHosts } from '../lib/api'
import { Host } from '../types'

interface Props {
  onSelectHost: (hostId: string) => void
}

export default function HostList({ onSelectHost }: Props) {
  const [hosts, setHosts] = useState<Host[]>([])
  const [filter, setFilter] = useState<string>("")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchHosts()
      .then(res => {
        setHosts(res.data.hosts || [])
        setLoading(false)
      })
      .catch(err => {
        setError(err.message)
        setLoading(false)
      })
  }, [])

  if (loading) return <div>Loading hosts...</div>
  if (error) return <div className="error">Error: {error}</div>

  const filteredHosts = hosts.filter(h => {
    const q = filter.trim().toLowerCase()
    if (!q) return true
    return (
      h.hostname.toLowerCase().includes(q) ||
      h.os_id.toLowerCase().includes(q) ||
      h.kernel.toLowerCase().includes(q)
    )
  })

  return (
    <div className="host-list">
      <h2>Monitored Hosts</h2>
      <div style={{ marginBottom: 12 }}>
        <input
          type="text"
          placeholder="Search hostname, os, kernel..."
          value={filter}
          onChange={e => setFilter(e.target.value)}
          style={{ padding: 8, width: '100%', maxWidth: 400 }}
        />
      </div>
      <table className="table">
        <thead>
          <tr>
            <th>Hostname</th>
            <th>OS</th>
            <th>Version</th>
            <th>Kernel</th>
            <th>Last Seen</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {filteredHosts.map(host => (
            <tr key={host.host_id}>
              <td>{host.hostname}</td>
              <td>{host.os_id}</td>
              <td>{host.os_version}</td>
              <td>{host.kernel}</td>
              <td>{host.last_seen ? new Date(host.last_seen).toLocaleString() : '-'}</td>
              <td>
                <button onClick={() => onSelectHost(host.host_id)} className="btn btn-primary">
                  Details
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {hosts.length === 0 && <p>No hosts found. Install the agent on a system.</p>}
    </div>
  )
}
