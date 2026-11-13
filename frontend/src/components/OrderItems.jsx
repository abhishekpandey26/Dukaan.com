import { formatPrice } from '../utils/format.js'

// Reused by the customer Orders page AND the admin Orders page (via props).
export default function OrderItems({ items }) {
  return (
    <ul className="order-items">
      {items.map((item, index) => (
        <li key={index}>
          <span>{item.quantity} × {item.product_name}</span>
          <span>{formatPrice(item.price_cents * item.quantity)}</span>
        </li>
      ))}
    </ul>
  )
}
