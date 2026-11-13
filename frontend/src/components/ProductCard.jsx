import { Link } from 'react-router-dom'
import { formatPrice } from '../utils/format.js'
import { PLACEHOLDER_IMAGE, onImageError } from '../utils/image.js'

// PROPS DRILLING example: this component doesn't know about the cart.
// The parent (Home) passes `product` and an `onAddToCart` function down as props.
export default function ProductCard({ product, onAddToCart, busy }) {
  const outOfStock = product.stock === 0

  return (
    <div className="card product-card">
      <Link to={`/products/${product.id}`}>
        <img src={product.image_url || PLACEHOLDER_IMAGE} alt={product.name} onError={onImageError} />
      </Link>
      <div className="card-body">
        <span className="tag">{product.category || 'General'}</span>
        <Link to={`/products/${product.id}`} className="product-name">{product.name}</Link>
        <div className="row">
          <strong>{formatPrice(product.price_cents)}</strong>
          <span className={outOfStock ? 'stock out' : 'stock'}>
            {outOfStock ? 'Out of stock' : `${product.stock} left`}
          </span>
        </div>
        <button className="btn btn-block" disabled={outOfStock || busy} onClick={() => onAddToCart(product.id)}>
          {busy ? 'Adding...' : 'Add to cart'}
        </button>
      </div>
    </div>
  )
}
