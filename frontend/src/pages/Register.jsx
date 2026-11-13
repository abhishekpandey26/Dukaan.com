import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'
import Message from '../components/Message.jsx'

export default function Register() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [form, setForm] = useState({ name: '', email: '', password: '' })
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleChange = (e) => setForm({ ...form, [e.target.name]: e.target.value })

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      await register(form.name, form.email, form.password)
      navigate('/')
    } catch (err) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="card form" onSubmit={handleSubmit}>
      <h2>Create account</h2>
      <Message text={error} />
      <label>Name<input name="name" value={form.name} onChange={handleChange} required /></label>
      <label>Email<input name="email" type="email" value={form.email} onChange={handleChange} required /></label>
      <label>Password<input name="password" type="password" minLength={6} value={form.password} onChange={handleChange} required /></label>
      <button className="btn btn-block" disabled={submitting}>{submitting ? 'Creating...' : 'Sign up'}</button>
      <p className="muted">Already have an account? <Link to="/login">Login</Link></p>
    </form>
  )
}
