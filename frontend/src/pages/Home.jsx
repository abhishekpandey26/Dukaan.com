import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api.js'
import { useAuth } from '../context/AuthContext.jsx'
import { useCart } from '../context/CartContext.jsx'
import { filterProducts, getCategories, paginate } from '../utils/search.js'
import ProductCard from '../components/ProductCard.jsx'
import Pagination from '../components/Pagination.jsx'
import Message from '../components/Message.jsx'

const PAGE_SIZE = 8

export default function Home() {
  const { user } = useAuth()
  const { addToCart } = useCart()
  const navigate = useNavigate()

  const [allProducts, setAllProducts] = useState([]) // loaded once from the API
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('')
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [busyId, setBusyId] = useState(null)
  const [message, setMessage] = useState({ type: 'error', text: '' })

  // 1. Load all products ONE time when the page opens.
  useEffect(() => {
    api.getAllProducts()
      .then(setAllProducts)
      .catch((err) => setMessage({ type: 'error', text: err.message }))
      .finally(() => setLoading(false))
  }, [])

  // 2. Search + filter in the browser. useMemo re-runs this only when
  //    the products, search text or category change (not on every render).
  const categories = useMemo(() => getCategories(allProducts), [allProducts])
  const filtered = useMemo(() => filterProducts(allProducts, search, category), [allProducts, search, category])

  // 3. Show only the current page.
  const totalPages = Math.ceil(filtered.length / PAGE_SIZE)
  const visible = paginate(filtered, page, PAGE_SIZE)

  // This function is passed DOWN to every ProductCard as a prop.
  async function handleAddToCart(productId) {
    if (!user) {
      navigate('/login')
      return
    }
    setBusyId(productId)
    try {
      await addToCart(productId, 1)
      setMessage({ type: 'success', text: 'Added to cart!' })
    } catch (err) {
      setMessage({ type: 'error', text: err.message })
    } finally {
      setBusyId(null)
    }
  }

  return (
    <>
      <div className="toolbar">
        <input
          type="search"
          placeholder="Search products..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }}
        />
        <select value={category} onChange={(e) => { setCategory(e.target.value); setPage(1) }}>
          <option value="">All categories</option>
          {categories.map((c) => <option key={c} value={c}>{c}</option>)}
        </select>
      </div>

      <Message type={message.type} text={message.text} />

      {!loading && (search || category) && (
        <p className="muted small">{filtered.length} {filtered.length === 1 ? 'product' : 'products'} found</p>
      )}

      {loading ? (
        <p className="empty">Loading products...</p>
      ) : visible.length === 0 ? (
        <p className="empty">No products match "{search}". Try another word or category.</p>
      ) : (
        <div className="grid">
          {visible.map((p) => (
            <ProductCard key={p.id} product={p} onAddToCart={handleAddToCart} busy={busyId === p.id} />
          ))}
        </div>
      )}

      <Pagination page={page} totalPages={totalPages} onChange={setPage} />
    </>
  )
}
