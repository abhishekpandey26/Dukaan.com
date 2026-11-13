import { useEffect, useState } from 'react'
import { api } from '../../api.js'
import { formatDate } from '../../utils/format.js'
import Message from '../../components/Message.jsx'

export default function AdminUsers() {
  const [users, setUsers] = useState([])
  const [error, setError] = useState('')

  useEffect(() => {
    api.admin.users().then(setUsers).catch((err) => setError(err.message))
  }, [])

  return (
    <>
      <h2>Users</h2>
      <Message text={error} />
      <div className="card">
        <table>
          <thead>
            <tr><th>ID</th><th>Name</th><th>Email</th><th>Role</th><th>Joined</th></tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id}>
                <td>{u.id}</td>
                <td>{u.name}</td>
                <td>{u.email}</td>
                <td><span className={`status ${u.role === 'admin' ? 'status-shipped' : ''}`}>{u.role}</span></td>
                <td>{formatDate(u.created_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  )
}
