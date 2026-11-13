import { createContext, useContext, useEffect, useState } from 'react'
import { api } from '../api.js'

// 1. Create the context (an empty "box" shared by the component tree).
const AuthContext = createContext(null)

// 2. The Provider owns the state and puts it in the box.
export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true) // true while we check a saved token

  // On first load: if we have a saved token, ask the backend who we are.
  useEffect(() => {
    if (!localStorage.getItem('token')) {
      setLoading(false)
      return
    }
    api.me()
      .then(setUser)
      .catch(() => localStorage.removeItem('token')) // token expired/invalid
      .finally(() => setLoading(false))
  }, [])

  async function login(email, password) {
    const data = await api.login(email, password)
    localStorage.setItem('token', data.token)
    setUser(data.user)
    return data.user
  }

  async function register(name, email, password) {
    const data = await api.register(name, email, password)
    localStorage.setItem('token', data.token)
    setUser(data.user)
    return data.user
  }

  function logout() {
    localStorage.removeItem('token')
    setUser(null)
  }

  const value = { user, loading, isAdmin: user?.role === 'admin', login, register, logout }
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// 3. A custom hook so components just write: const { user } = useAuth()
export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used inside <AuthProvider>')
  return context
}
