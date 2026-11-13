import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'
import Message from '../components/Message.jsx'

export default function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  // One state object for the whole form ("controlled inputs").
  const [form, setForm] = useState({ email: '', password: '' })
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleChange = (e) => setForm({ ...form, [e.target.name]: e.target.value })

  async function handleSubmit(e) {
    e.preventDefault() // stop the browser from reloading the page
    setError('')
    setSubmitting(true)
    try {
      const user = await login(form.email, form.password)
      // Go back where the user came from, or to admin/home.
      navigate(location.state?.from || (user.role === 'admin' ? '/admin' : '/'))
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="card form" onSubmit={handleSubmit}>
      <h2>Login</h2>
      <Message text={error} />
      <label>Email<input name="email" type="email" value={form.email} onChange={handleChange} required /></label>
      <label>Password<input name="password" type="password" value={form.password} onChange={handleChange} required /></label>
      <button className="btn btn-block" disabled={submitting}>{submitting ? 'Logging in...' : 'Login'}</button>
      <p className="muted">No account? <Link to="/register">Sign up</Link></p>
      <p className="muted small">Admin demo: admin@gocommerce.com / admin123</p>
    </form>
  )
}
