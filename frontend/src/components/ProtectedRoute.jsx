import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'

// Wrap a page with this to require login (and optionally admin role).
// NOTE: this only hides pages in the UI. The real security is in the Go
// middleware: the API refuses the request without a valid admin token.
export default function ProtectedRoute({ children, adminOnly = false }) {
  const { user, loading, isAdmin } = useAuth()
  const location = useLocation()

  if (loading) return <p className="empty">Loading...</p>
  if (!user) return <Navigate to="/login" replace state={{ from: location.pathname }} />
  if (adminOnly && !isAdmin) return <Navigate to="/" replace />
  return children
}
