import { useEffect, useState } from 'react'
import { api } from '../../api.js'
import { formatPrice } from '../../utils/format.js'
import Message from '../../components/Message.jsx'

// StatCard gets everything it needs from props.
function StatCard({ label, value, highlight }) {
  return (
    <div className={`card stat ${highlight ? 'stat-warn' : ''}`}>
      <span className="muted">{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

export default function Dashboard() {
  const [stats, setStats] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.admin.stats().then(setStats).catch((err) => setError(err.message))
  }, [])

  if (!stats) return <><Message text={error} /><p className="empty">Loading...</p></>

  return (
    <>
      <h2>Dashboard</h2>
      <div className="stats">
        <StatCard label="Revenue" value={formatPrice(stats.revenue_cents)} />
        <StatCard label="Orders" value={stats.orders} />
        <StatCard label="Pending orders" value={stats.pending_orders} highlight={stats.pending_orders > 0} />
        <StatCard label="Products" value={stats.products} />
        <StatCard label="Low stock (< 5)" value={stats.low_stock} highlight={stats.low_stock > 0} />
        <StatCard label="Users" value={stats.users} />
      </div>
    </>
  )
}
