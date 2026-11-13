import { createContext, useContext, useEffect, useState } from 'react'
import { api } from '../api.js'
import { useAuth } from './AuthContext.jsx'

const CartContext = createContext(null)
const EMPTY_CART = { items: [], total_cents: 0 }

// The cart lives on the server. This context just keeps a copy so that
// the Navbar badge, Cart page, etc. all show the same data.
export function CartProvider({ children }) {
  const { user } = useAuth()
  const [cart, setCart] = useState(EMPTY_CART)

  // Re-load the cart whenever the logged-in user changes (login / logout).
  useEffect(() => {
    if (!user) {
      setCart(EMPTY_CART)
      return
    }
    api.getCart().then(setCart).catch(() => setCart(EMPTY_CART))
  }, [user])

  // Every cart endpoint returns the updated cart, so we just store the response.
  const addToCart = async (productId, quantity = 1) => setCart(await api.addToCart(productId, quantity))
  const updateQuantity = async (productId, quantity) => setCart(await api.updateCartItem(productId, quantity))
  const removeItem = async (productId) => setCart(await api.removeCartItem(productId))
  const clearCart = async () => setCart(await api.clearCart())

  async function checkout() {
    const order = await api.placeOrder()
    setCart(EMPTY_CART) // the backend emptied the cart inside the transaction
    return order
  }

  const itemCount = cart.items.reduce((sum, item) => sum + item.quantity, 0)

  const value = { cart, itemCount, addToCart, updateQuantity, removeItem, clearCart, checkout }
  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}

export function useCart() {
  const context = useContext(CartContext)
  if (!context) throw new Error('useCart must be used inside <CartProvider>')
  return context
}
