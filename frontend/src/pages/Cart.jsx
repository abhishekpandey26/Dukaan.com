import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useCart } from '../context/CartContext.jsx'
import { formatPrice } from '../utils/format.js'
import { PLACEHOLDER_IMAGE, onImageError } from '../utils/image.js'
import Message from '../components/Message.jsx'

export default function Cart() {
  const { cart, updateQuantity, removeItem, clearCart, checkout } = useCart()
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [placing, setPlacing] = useState(false)

  // Small helper so every action shows errors the same way.
  async function run(action) {
    setError('')
    try {
      await action()
    } catch (err) {
      setError(err.message)
    }
  }

  async function handleCheckout() {
    setPlacing(true)
    await run(async () => {
      const order = await checkout()
      navigate('/orders', { state: { message: `Order #${order.id} placed successfully!` } })
    })
    setPlacing(false)
  }

  if (cart.items.length === 0) {
    return <p className="empty">Your cart is empty. <Link to="/">Go shopping</Link></p>
  }

  return (
    <>
      <h2>Your cart</h2>
      <Message text={error} />
      <div className="card">
        <table>
          <thead>
            <tr><th>Product</th><th>Price</th><th>Quantity</th><th>Subtotal</th><th></th></tr>
          </thead>
          <tbody>
            {cart.items.map((item) => (
              <tr key={item.product_id}>
                <td className="cell-product">
                  <img src={item.image_url || PLACEHOLDER_IMAGE} alt="" onError={onImageError} />
                  <Link to={`/products/${item.product_id}`}>{item.name}</Link>
                </td>
                <td>{formatPrice(item.price_cents)}</td>
                <td>
                  <div className="qty">
                    <button className="btn btn-small btn-outline" disabled={item.quantity <= 1}
                      onClick={() => run(() => updateQuantity(item.product_id, item.quantity - 1))}>-</button>
                    <span>{item.quantity}</span>
                    <button className="btn btn-small btn-outline" disabled={item.quantity >= item.stock}
                      onClick={() => run(() => updateQuantity(item.product_id, item.quantity + 1))}>+</button>
                  </div>
                </td>
                <td>{formatPrice(item.subtotal_cents)}</td>
                <td>
                  <button className="btn btn-small btn-danger" onClick={() => run(() => removeItem(item.product_id))}>Remove</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        <div className="cart-footer">
          <button className="btn btn-outline" onClick={() => run(clearCart)}>Clear cart</button>
          <div className="row">
            <strong className="price">Total: {formatPrice(cart.total_cents)}</strong>
            <button className="btn" onClick={handleCheckout} disabled={placing}>
              {placing ? 'Placing order...' : 'Place order'}
            </button>
          </div>
        </div>
      </div>
    </>
  )
}
