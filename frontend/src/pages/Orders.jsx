import { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { api } from '../api.js'
import { formatDate, formatPrice } from '../utils/format.js'
import Message from '../components/Message.jsx'
import OrderItems from '../components/OrderItems.jsx'

export default function Orders() {
  const location = useLocation()
  const [orders, setOrders] = useState(null)
  const [message, setMessage] = useState({ type: 'success', text: location.state?.message || '' })

  useEffect(() => {
    api.getOrders()
      .then(setOrders)
      .catch((err) => setMessage({ type: 'error', text: err.message }))
  }, [])

  async function handleCancel(id) {
    if (!confirm(`Cancel order #${id}?`)) return
    try {
      const updated = await api.cancelOrder(id)
      // Replace just the changed order in the list (immutable update).
      setOrders(orders.map((o) => (o.id === id ? updated : o)))
      setMessage({ type: 'success', text: `Order #${id} cancelled. Items were returned to stock.` })
    } catch (err) {
      setMessage({ type: 'error', text: err.message })
    }
  }

  if (!orders) return <><Message {...message} /><p className="empty">Loading...</p></>

  return (
    <>
      <h2>My orders</h2>
      <Message {...message} />
      {orders.length === 0 && <p className="empty">You have no orders yet.</p>}
      {orders.map((order) => (
        <div key={order.id} className="card order">
          <div className="order-header">
            <div>
              <strong>Order #{order.id}</strong>
              <span className="muted"> · {formatDate(order.created_at)}</span>
            </div>
            <div className="row">
              <span className={`status status-${order.status}`}>{order.status}</span>
              {order.status === 'pending' && (
                <button className="btn btn-small btn-danger" onClick={() => handleCancel(order.id)}>Cancel</button>
              )}
            </div>
          </div>
          <OrderItems items={order.items} />
          <p className="order-total">Total: <strong>{formatPrice(order.total_cents)}</strong></p>
        </div>
      ))}
    </>
  )
}
