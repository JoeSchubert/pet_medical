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

  if (loading) return <div className="page"><p role="status">Loading…</p></div>
  if (error) return <div className="page"><p className="error" role="alert">{error}</p></div>

  return (
    <div className="page" aria-label="User management">
      <header className="page-header">
        <h1 id="users-title">User management</h1>
      </header>
      <p className="muted">Manage users and roles. Only admins can access this page.</p>

      <div className="users-layout">
      <section className="section card" aria-labelledby="users-add-heading">
        <h2 id="users-add-heading">Add user</h2>
        <form onSubmit={addUser} className="form-inline" aria-describedby={addError ? 'users-add-error' : undefined}>
          <input
            id="users-add-display-name"
            type="text"
            placeholder={t('common.displayName')}
            value={addDisplayName}
            onChange={(e) => setAddDisplayName(e.target.value)}
            className="input"
            autoComplete="name"
            aria-label={t('common.displayName')}
          />
          <input
            id="users-add-email"
            type="email"
            placeholder="Email"
            value={addEmail}
            onChange={(e) => setAddEmail(e.target.value)}
            className="input"
            autoComplete="email"
            aria-label="Email"
          />
          <input
            id="users-add-password"
            type="password"
            placeholder="Password"
            value={addPassword}
            onChange={(e) => setAddPassword(e.target.value)}
            className="input"
            autoComplete="new-password"
            aria-label="Password"
          />
          <select id="users-add-role" value={addRole} onChange={(e) => setAddRole(e.target.value)} className="input" aria-label="Role">
            <option value="user">user</option>
            <option value="admin">admin</option>
          </select>
          <button type="submit" className="btn btn-primary" disabled={addSubmitting} aria-busy={addSubmitting}>
            {addSubmitting ? 'Adding…' : 'Add user'}
          </button>
        </form>
        {addError && <p id="users-add-error" className="error" role="alert">{addError}</p>}
      </section>

      <div className="section table-wrap users-table-wrap" role="region" aria-labelledby="users-title">
        <table className="data-table" aria-label="Users list">
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
                  <Link to={`/settings/${u.id}`} className="btn btn-sm btn-secondary" title="Settings" aria-label={`Settings for ${u.display_name}`}>
                    <Icon icon="mdi:cog" width={16} height={16} aria-hidden />
                    Settings
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      </div>
    </div>
  )
}
