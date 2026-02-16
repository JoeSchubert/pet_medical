import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Icon } from '@iconify/react'
import { useTranslation } from '../i18n/context'
import { usersApi, type UserListEntry } from '../api/client'

export default function Users() {
  const { t } = useTranslation()
  const [users, setUsers] = useState<UserListEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [addDisplayName, setAddDisplayName] = useState('')
  const [addEmail, setAddEmail] = useState('')
  const [addPassword, setAddPassword] = useState('')
  const [addRole, setAddRole] = useState('user')
  const [addSubmitting, setAddSubmitting] = useState(false)
  const [addError, setAddError] = useState('')

  function load() {
    setLoading(true)
    usersApi
      .list()
      .then(setUsers)
      .catch(() => setError('Failed to load users'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  async function addUser(e: React.FormEvent) {
    e.preventDefault()
    setAddError('')
    if (!addDisplayName.trim() || !addEmail.trim() || !addPassword) {
      setAddError(t('users.displayNameEmailPasswordRequired'))
      return
    }
    setAddSubmitting(true)
    try {
      await usersApi.create(addDisplayName.trim(), addEmail.trim(), addPassword, addRole)
      setAddDisplayName('')
      setAddEmail('')
      setAddPassword('')
      setAddRole('user')
      load()
    } catch (err) {
      setAddError(err instanceof Error ? err.message : 'Failed to create user')
    } finally {
      setAddSubmitting(false)
    }
  }

  if (loading) return <div className="page"><p>Loading…</p></div>
  if (error) return <div className="page"><p className="error">{error}</p></div>

  return (
    <div className="page">
      <header className="page-header">
        <h1>User management</h1>
      </header>
      <p className="muted">Manage users and roles. Only admins can access this page.</p>

      <section className="section card">
        <h2>Add user</h2>
        <form onSubmit={addUser} className="form-inline">
          <input
            type="text"
            placeholder={t('common.displayName')}
            value={addDisplayName}
            onChange={(e) => setAddDisplayName(e.target.value)}
            className="input"
            autoComplete="name"
          />
          <input
            type="email"
            placeholder="Email"
            value={addEmail}
            onChange={(e) => setAddEmail(e.target.value)}
            className="input"
            autoComplete="email"
          />
          <input
            type="password"
            placeholder="Password"
            value={addPassword}
            onChange={(e) => setAddPassword(e.target.value)}
            className="input"
            autoComplete="new-password"
          />
          <select value={addRole} onChange={(e) => setAddRole(e.target.value)} className="input">
            <option value="user">user</option>
            <option value="admin">admin</option>
          </select>
          <button type="submit" className="btn btn-primary" disabled={addSubmitting}>
            {addSubmitting ? 'Adding…' : 'Add user'}
          </button>
        </form>
        {addError && <p className="error">{addError}</p>}
      </section>

      <div className="section table-wrap">
        <table className="data-table">
          <thead>
            <tr>
              <th>{t('common.displayName')}</th>
              <th>Email</th>
              <th>Role</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id}>
                <td>{u.display_name}</td>
                <td>{u.email || '—'}</td>
                <td>{u.role}</td>
                <td>{new Date(u.created_at).toLocaleDateString()}</td>
                <td>
                  <Link to={`/settings/${u.id}`} className="btn btn-sm btn-secondary" title="Settings">
                    <Icon icon="mdi:cog" width={16} height={16} />
                    Settings
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
