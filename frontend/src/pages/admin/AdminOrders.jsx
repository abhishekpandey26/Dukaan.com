import { useEffect, useState } from 'react'
import { api } from '../../api.js'
import { formatDate, formatPrice, ORDER_STATUSES } from '../../utils/format.js'
import Message from '../../components/Message.jsx'
import OrderItems from '../../components/OrderItems.jsx'

export default function AdminOrders() {
  const [orders, setOrders] = useState([])
  const [filter, setFilter] = useState('')
  const [message, setMessage] = useState({ type: 'error', text: '' })

  useEffect(() => {
    api.admin.orders(filter)
      .then(setOrders)
      .catch((err) => setMessage({ type: 'error', text: err.message }))
  }, [filter])

  async function handleStatusChange(order, status) {
    if (status === 'cancelled' && !confirm(`Cancel order #${order.id}? Stock will be restored.`)) return
    try {
      const updated = await api.admin.updateOrderStatus(order.id, status)
      setOrders(orders.map((o) => (o.id === order.id ? updated : o)))
      setMessage({ type: 'success', text: `Order #${order.id} is now "${status}".` })
    } catch (err) {
      setMessage({ type: 'error', text: err.message })
    }
  }

  return (
    <>
      <div className="row space-between">
        <h2>Orders</h2>
        <select value={filter} onChange={(e) => setFilter(e.target.value)}>
          <option value="">All statuses</option>
          {ORDER_STATUSES.map((s) => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>

      <Message {...message} />
      {orders.length === 0 && <p className="empty">No orders.</p>}

      {orders.map((order) => (
        <div key={order.id} className="card order">
          <div className="order-header">
            <div>
              <strong>Order #{order.id}</strong>
              <span className="muted"> · {order.user_email} · {formatDate(order.created_at)}</span>
            </div>
            <select
              value={order.status}
              disabled={order.status === 'cancelled'}
              onChange={(e) => handleStatusChange(order, e.target.value)}
            >
              {ORDER_STATUSES.map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
          <OrderItems items={order.items} />
          <p className="order-total">Total: <strong>{formatPrice(order.total_cents)}</strong></p>
        </div>
      ))}
    </>
  )
}
