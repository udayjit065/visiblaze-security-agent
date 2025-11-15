import { useState } from 'react'
import HostDetail from './components/HostDetail'
import HostList from './components/HostList'
import './styles.css'

export default function App() {
  const [selectedHostId, setSelectedHostId] = useState<string | null>(null)

  return (
    <div className="app">
      <header className="header">
        <h1>Visiblaze Security Agent</h1>
        <p>CIS Compliance & Package Inventory</p>
      </header>
      <div className="container">
        {!selectedHostId ? (
          <HostList onSelectHost={setSelectedHostId} />
        ) : (
          <HostDetail hostId={selectedHostId} onBack={() => setSelectedHostId(null)} />
        )}
      </div>
    </div>
  )
}
