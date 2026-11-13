import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'
import { useCart } from '../context/CartContext.jsx'

export default function Navbar() {
  const { user, isAdmin, logout } = useAuth()
  const { itemCount } = useCart()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/')
  }

  return (
    <header className="navbar">
      <div className="container navbar-inner">
        <Link to="/" className="brand">GoCommerce</Link>
        <nav className="nav-links">
          <NavLink to="/" end>Shop</NavLink>
          {user && <NavLink to="/orders">My orders</NavLink>}
          {isAdmin && <NavLink to="/admin">Admin</NavLink>}
          {user && (
            <NavLink to="/cart">
              Cart {itemCount > 0 && <span className="badge">{itemCount}</span>}
            </NavLink>
          )}
          {user ? (
            <>
              <span className="muted">Hi, {user.name}</span>
              <button className="btn btn-small btn-outline" onClick={handleLogout}>Logout</button>
            </>
          ) : (
            <>
              <NavLink to="/login">Login</NavLink>
              <Link to="/register" className="btn btn-small">Sign up</Link>
            </>
          )}
        </nav>
      </div>
    </header>
  )
}
