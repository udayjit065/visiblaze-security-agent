import { useEffect, useState } from 'react'
import { fetchPackages } from '../lib/api'
import { Package } from '../types'

interface Props {
  hostId: string
}

export default function PackagesTable({ hostId }: Props) {
  const [packages, setPackages] = useState<Package[]>([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchPackages(hostId)
      .then(res => {
        setPackages(res.data.packages || [])
        setLoading(false)
      })
      .catch(err => {
        console.error(err)
        setLoading(false)
      })
  }, [hostId])

  const filtered = packages.filter(p =>
    p.name.toLowerCase().includes(search.toLowerCase())
  )

  if (loading) return <div>Loading packages...</div>

  return (
    <div className="packages-table">
      <input
        type="text"
        placeholder="Search packages..."
        value={search}
        onChange={e => setSearch(e.target.value)}
        className="search-input"
      />
      <p>{filtered.length} / {packages.length} packages</p>
      <table className="table">
        <thead>
          <tr>
            <th>Package</th>
            <th>Version</th>
            <th>Arch</th>
            <th>Manager</th>
          </tr>
        </thead>
        <tbody>
          {filtered.slice(0, 50).map(pkg => (
            <tr key={pkg.name + pkg.arch}>
              <td>{pkg.name}</td>
              <td>{pkg.version}</td>
              <td>{pkg.arch}</td>
              <td>{pkg.manager}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
