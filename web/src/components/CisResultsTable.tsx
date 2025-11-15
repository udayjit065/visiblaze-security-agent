import React, { useEffect, useState } from 'react'
import { fetchCISResults } from '../lib/api'
import { CISResult } from '../types'
import './CisResultsTable.css'

interface Props {
  hostId: string
}

export default function CisResultsTable({ hostId }: Props) {
  const [results, setResults] = useState<CISResult[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedCheck, setExpandedCheck] = useState<string | null>(null)

  useEffect(() => {
    fetchCISResults(hostId)
      .then(res => {
        setResults(res.data.cis_results || [])
        setLoading(false)
      })
      .catch(err => {
        console.error(err)
        setLoading(false)
      })
  }, [hostId])

  if (loading) return <div>Loading CIS results...</div>

  const passCount = results.filter(r => r.status === 'pass').length
  const failCount = results.filter(r => r.status === 'fail').length

  return (
    <div className="cis-results">
      <div className="cis-summary">
        <div className="summary-item pass">
          <strong>{passCount}</strong> <span>Passed</span>
        </div>
        <div className="summary-item fail">
          <strong>{failCount}</strong> <span>Failed</span>
        </div>
        <div className="summary-item manual">
          <strong>{results.length - passCount - failCount}</strong> <span>Manual</span>
        </div>
      </div>

      <table className="table cis-table">
        <thead>
          <tr>
            <th>Check ID</th>
            <th>Title</th>
            <th>Status</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {results.map(result => (
            <React.Fragment key={result.check_id}>
              <tr>
                <td><strong>{result.check_id}</strong></td>
                <td>{result.title}</td>
                <td>
                  <span className={`status-badge ${result.status}`}>
                    {result.status.toUpperCase()}
                  </span>
                </td>
                <td>
                  <button
                    onClick={() => setExpandedCheck(
                      expandedCheck === result.check_id ? null : result.check_id
                    )}
                    className="btn btn-sm"
                  >
                    {expandedCheck === result.check_id ? '▼ Hide' : '▶ Show'} Evidence
                  </button>
                </td>
              </tr>
              {expandedCheck === result.check_id && (
                <tr className="evidence-row">
                  <td colSpan={4}>
                    <pre>{JSON.stringify(result.evidence, null, 2)}</pre>
                  </td>
                </tr>
              )}
            </React.Fragment>
          ))}
        </tbody>
      </table>
    </div>
  )
}
