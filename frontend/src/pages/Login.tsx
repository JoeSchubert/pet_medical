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
    <div className="login-page" role="main" aria-label="Sign in">
      <div className="login-wrap">
        <div className="login-card">
          <h1 id="login-title">{t('nav.brand')}</h1>
          <p className="tagline" id="login-tagline">Your pet's health portfolio</p>
          <form onSubmit={handleSubmit} aria-labelledby="login-title" aria-describedby={error ? 'login-error' : undefined}>
            {error && (
              <div id="login-error" className="error" role="alert" aria-live="assertive">
                {error}
              </div>
            )}
            <label htmlFor="login-email">
              {t('login.email')}
              <div className="input-wrap">
                <Icon icon="mdi:email-outline" className="input-icon" width={24} height={24} aria-hidden />
                <input
                  id="login-email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                  placeholder={t('login.email')}
                  required
                  aria-invalid={!!error}
                  aria-describedby={error ? 'login-error' : undefined}
                />
              </div>
            </label>
            <label htmlFor="login-password">
              {t('login.password')}
              <div className="input-wrap">
                <Icon icon="mdi:lock-outline" className="input-icon" width={24} height={24} aria-hidden />
                <input
                  id="login-password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="current-password"
                  placeholder={t('login.password')}
                  required
                  aria-invalid={!!error}
                  aria-describedby={error ? 'login-error' : undefined}
                />
              </div>
            </label>
            <button id="login-submit" type="submit" disabled={submitting} aria-busy={submitting}>
              {submitting ? (
                t('login.signingIn')
              ) : (
                <>
                  <Icon icon="mdi:login" width={20} height={20} aria-hidden />
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
