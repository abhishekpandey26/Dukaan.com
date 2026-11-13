import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../api.js'
import { useAuth } from '../context/AuthContext.jsx'
import { useCart } from '../context/CartContext.jsx'
import { formatPrice } from '../utils/format.js'
import { PLACEHOLDER_IMAGE, onImageError } from '../utils/image.js'
import Message from '../components/Message.jsx'

export default function ProductDetail() {
  const { id } = useParams() // the :id from the URL /products/:id
  const { user } = useAuth()
  const { addToCart } = useCart()
  const navigate = useNavigate()

  const [product, setProduct] = useState(null)
  const [quantity, setQuantity] = useState(1)
  const [message, setMessage] = useState({ type: 'error', text: '' })

  useEffect(() => {
    api.getProduct(id)
      .then(setProduct)
      .catch((err) => setMessage({ type: 'error', text: err.message }))
  }, [id])

  async function handleAdd() {
    if (!user) return navigate('/login', { state: { from: `/products/${id}` } })
    try {
      await addToCart(product.id, quantity)
      setMessage({ type: 'success', text: `Added ${quantity} to cart.` })
    } catch (err) {
      setMessage({ type: 'error', text: err.message })
    }
  }

  if (!product) return <><Message {...message} /><p className="empty">Loading...</p></>

  return (
    <>
      <Link to="/">&larr; Back to shop</Link>
      <div className="detail card">
        <img src={product.image_url || PLACEHOLDER_IMAGE} alt={product.name} onError={onImageError} />
        <div className="detail-info">
          <span className="tag">{product.category || 'General'}</span>
          <h1>{product.name}</h1>
          <p className="price">{formatPrice(product.price_cents)}</p>
          <p>{product.description}</p>
          <p className={product.stock === 0 ? 'stock out' : 'stock'}>
            {product.stock === 0 ? 'Out of stock' : `${product.stock} in stock`}
          </p>
          {product.stock > 0 && (
            <div className="row">
              <input
                type="number"
                min="1"
                max={product.stock}
                value={quantity}
                onChange={(e) => setQuantity(Math.max(1, Number(e.target.value)))}
                style={{ width: 80 }}
              />
              <button className="btn" onClick={handleAdd}>Add to cart</button>
            </div>
          )}
          <Message {...message} />
        </div>
      </div>
    </>
  )
}
