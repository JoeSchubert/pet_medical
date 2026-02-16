import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Icon } from '@iconify/react'
import { useAuth } from '../contexts/AuthContext'
import { useTranslation } from '../i18n/context'
import { logAuth } from '../lib/log'

export default function Login() {
  const { login } = useAuth()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    logAuth('Login handleSubmit email=', email)
    setError('')
    setSubmitting(true)
    try {
      await login(email, password)
      logAuth('Login handleSubmit login() succeeded, calling navigate(/)')
      navigate('/', { replace: true })
      logAuth('Login handleSubmit navigate() returned')
    } catch (err) {
      logAuth('Login handleSubmit error', err)
      const msg = err instanceof Error ? err.message : ''
      setError(msg && msg.startsWith('error.') ? t(msg) : t('login.failed'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-wrap">
        <div className="login-card">
          <h1>{t('nav.brand')}</h1>
          <p className="tagline">Your pet's health portfolio</p>
          <form onSubmit={handleSubmit}>
            {error && <div className="error">{error}</div>}
            <label>
              {t('login.email')}
              <div className="input-wrap">
                <Icon icon="mdi:email-outline" className="input-icon" width={24} height={24} />
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                  placeholder={t('login.email')}
                  required
                />
              </div>
            </label>
            <label>
              {t('login.password')}
              <div className="input-wrap">
                <Icon icon="mdi:lock-outline" className="input-icon" width={24} height={24} />
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="current-password"
                  placeholder={t('login.password')}
                  required
                />
              </div>
            </label>
            <button type="submit" disabled={submitting}>
              {submitting ? (
                t('login.signingIn')
              ) : (
                <>
                  <Icon icon="mdi:login" width={20} height={20} />
                  {t('login.submit')}
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  )
}
