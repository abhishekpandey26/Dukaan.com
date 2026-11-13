// All calls to the Go backend live in this one file.
// Components never call fetch() directly; they call api.something().

const BASE_URL = import.meta.env.VITE_API_URL || '/api'

async function request(path, { method = 'GET', body } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  const token = localStorage.getItem('token')
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(BASE_URL + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  if (res.status === 204) return null // "No Content" (e.g. after DELETE)

  const data = await res.json().catch(() => null)
  if (!res.ok) {
    // Our Go API always sends errors as {"error": "..."}
    throw new Error(data?.error || `Request failed (${res.status})`)
  }
  return data
}

export const api = {
  // Auth
  login: (email, password) => request('/auth/login', { method: 'POST', body: { email, password } }),
  register: (name, email, password) => request('/auth/register', { method: 'POST', body: { name, email, password } }),
  me: () => request('/auth/me'),

  // Products (public)
  getProducts: (params = {}) => request('/products?' + new URLSearchParams(params)),
  // Loads EVERY product (the API returns max 100 per page, so we loop over pages).
  // Used for client-side search.
  getAllProducts: async () => {
    const all = []
    for (let page = 1; ; page++) {
      const data = await request(`/products?page=${page}&limit=100`)
      all.push(...data.products)
      if (all.length >= data.total || data.products.length === 0) return all
    }
  },
  getProduct: (id) => request(`/products/${id}`),
  getCategories: () => request('/categories'),

  // Cart
  getCart: () => request('/cart'),
  addToCart: (productId, quantity) => request('/cart', { method: 'POST', body: { product_id: productId, quantity } }),
  updateCartItem: (productId, quantity) => request(`/cart/${productId}`, { method: 'PUT', body: { quantity } }),
  removeCartItem: (productId) => request(`/cart/${productId}`, { method: 'DELETE' }),
  clearCart: () => request('/cart', { method: 'DELETE' }),

  // Orders
  placeOrder: () => request('/orders', { method: 'POST' }),
  getOrders: () => request('/orders'),
  cancelOrder: (id) => request(`/orders/${id}/cancel`, { method: 'POST' }),

  // Admin
  admin: {
    stats: () => request('/admin/stats'),
    createProduct: (product) => request('/admin/products', { method: 'POST', body: product }),
    updateProduct: (id, product) => request(`/admin/products/${id}`, { method: 'PUT', body: product }),
    deleteProduct: (id) => request(`/admin/products/${id}`, { method: 'DELETE' }),
    orders: (status = '') => request('/admin/orders' + (status ? `?status=${status}` : '')),
    updateOrderStatus: (id, status) => request(`/admin/orders/${id}/status`, { method: 'PUT', body: { status } }),
    users: () => request('/admin/users'),
  },
}
