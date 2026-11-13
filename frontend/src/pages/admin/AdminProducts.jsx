import { useEffect, useMemo, useState } from 'react'
import { api } from '../../api.js'
import { formatPrice } from '../../utils/format.js'
import { filterProducts, paginate } from '../../utils/search.js'
import Message from '../../components/Message.jsx'
import Pagination from '../../components/Pagination.jsx'
import ProductForm from './ProductForm.jsx'

const PAGE_SIZE = 10

export default function AdminProducts() {
  const [allProducts, setAllProducts] = useState([])
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [editing, setEditing] = useState(undefined) // undefined = form closed, null = new, object = edit
  const [message, setMessage] = useState({ type: 'error', text: '' })

  // Reload the full list (after create / update / delete).
  function loadProducts() {
    api.getAllProducts()
      .then(setAllProducts)
      .catch((err) => setMessage({ type: 'error', text: err.message }))
  }

  useEffect(() => {
    loadProducts()
  }, [])

  // Search happens in the browser, same helper as the shop page.
  const filtered = useMemo(() => filterProducts(allProducts, search), [allProducts, search])
  const visible = paginate(filtered, page, PAGE_SIZE)

  // Passed to <ProductForm onSubmit={...}>.
  async function handleSave(data) {
    try {
      if (editing) {
        await api.admin.updateProduct(editing.id, data)
        setMessage({ type: 'success', text: `Product "${data.name}" updated.` })
      } else {
        await api.admin.createProduct(data)
        setMessage({ type: 'success', text: `Product "${data.name}" created.` })
      }
      setEditing(undefined)
      loadProducts()
    } catch (err) {
      setMessage({ type: 'error', text: err.message })
    }
  }

  async function handleDelete(product) {
    if (!confirm(`Delete "${product.name}"? This cannot be undone.`)) return
    try {
      await api.admin.deleteProduct(product.id)
      setMessage({ type: 'success', text: `Product "${product.name}" deleted.` })
      loadProducts()
    } catch (err) {
      setMessage({ type: 'error', text: err.message })
    }
  }

  return (
    <>
      <div className="row space-between">
        <h2>Products</h2>
        {editing === undefined && <button className="btn" onClick={() => setEditing(null)}>+ Add product</button>}
      </div>

      <Message {...message} />

      {editing !== undefined && (
        <ProductForm
          key={editing?.id ?? 'new'} // new key = fresh form state when switching products
          initialProduct={editing}
          onSubmit={handleSave}
          onCancel={() => setEditing(undefined)}
        />
      )}

      <div className="toolbar">
        <input type="search" placeholder="Search by name, description or category..." value={search}
          onChange={(e) => { setSearch(e.target.value); setPage(1) }} />
      </div>

      <div className="card">
        <table>
          <thead>
            <tr><th>ID</th><th>Name</th><th>Category</th><th>Price</th><th>Stock</th><th>Actions</th></tr>
          </thead>
          <tbody>
            {visible.map((p) => (
              <tr key={p.id}>
                <td>{p.id}</td>
                <td>{p.name}</td>
                <td>{p.category}</td>
                <td>{formatPrice(p.price_cents)}</td>
                <td className={p.stock < 5 ? 'stock out' : ''}>{p.stock}</td>
                <td className="actions">
                  <button className="btn btn-small btn-outline" onClick={() => setEditing(p)}>Edit</button>
                  <button className="btn btn-small btn-danger" onClick={() => handleDelete(p)}>Delete</button>
                </td>
              </tr>
            ))}
            {visible.length === 0 && <tr><td colSpan="6" className="empty">No products found.</td></tr>}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={Math.ceil(filtered.length / PAGE_SIZE)} onChange={setPage} />
    </>
  )
}
