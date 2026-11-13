import { useState } from 'react'

const EMPTY = { name: '', description: '', price: '', stock: '', category: '', image_url: '' }

// Used for BOTH "create" and "edit". The parent decides via props:
//   initialProduct -> null for create, a product object for edit
//   onSubmit       -> function the parent runs with the form data
//   onCancel       -> function to close the form
export default function ProductForm({ initialProduct, onSubmit, onCancel }) {
  const [form, setForm] = useState(
    initialProduct
      ? { ...initialProduct, price: (initialProduct.price_cents / 100).toFixed(2), stock: String(initialProduct.stock) }
      : EMPTY,
  )
  const [saving, setSaving] = useState(false)

  const handleChange = (e) => setForm({ ...form, [e.target.name]: e.target.value })

  async function handleSubmit(e) {
    e.preventDefault()
    setSaving(true)
    // The admin types dollars; the API wants integer cents.
    await onSubmit({
      name: form.name,
      description: form.description,
      price_cents: Math.round(Number(form.price) * 100),
      stock: Number(form.stock),
      category: form.category,
      image_url: form.image_url,
    })
    setSaving(false)
  }

  return (
    <form className="card form form-wide" onSubmit={handleSubmit}>
      <h3>{initialProduct ? `Edit product #${initialProduct.id}` : 'New product'}</h3>
      <label>Name<input name="name" value={form.name} onChange={handleChange} required /></label>
      <label>Description<textarea name="description" rows="3" value={form.description} onChange={handleChange} /></label>
      <div className="form-row">
        <label>Price ($)<input name="price" type="number" step="0.01" min="0" value={form.price} onChange={handleChange} required /></label>
        <label>Stock<input name="stock" type="number" min="0" value={form.stock} onChange={handleChange} required /></label>
        <label>Category<input name="category" value={form.category} onChange={handleChange} /></label>
      </div>
      <label>Image URL<input name="image_url" value={form.image_url} onChange={handleChange} placeholder="https://..." /></label>
      <div className="row">
        <button className="btn" disabled={saving}>{saving ? 'Saving...' : 'Save'}</button>
        <button type="button" className="btn btn-outline" onClick={onCancel}>Cancel</button>
      </div>
    </form>
  )
}
