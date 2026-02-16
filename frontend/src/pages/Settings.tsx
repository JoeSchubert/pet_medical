import { Icon } from '@iconify/react'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { usePWAInstall } from '../contexts/PWAInstallContext'
import { useTranslation } from '../i18n/context'
import { settingsApi, changePassword as apiChangePassword, type Settings, type WeightUnit } from '../api/client'

export default function Settings() {
  const { user, refreshUser } = useAuth()
  const { t } = useTranslation()
  const pwa = usePWAInstall()
  const { userId } = useParams<{ userId?: string }>()
  const isAdminEditing = !!userId && user?.role === 'admin'
  const [settings, setSettings] = useState<Settings | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [passwordCurrent, setPasswordCurrent] = useState('')
  const [passwordNew, setPasswordNew] = useState('')
  const [passwordConfirm, setPasswordConfirm] = useState('')
  const [passwordChanging, setPasswordChanging] = useState(false)
  const [passwordError, setPasswordError] = useState('')
  const [passwordMessage, setPasswordMessage] = useState('')

  useEffect(() => {
    const normalize = (s: Settings) => ({
      weight_unit: s.weight_unit ?? 'lbs',
      currency: s.currency ?? 'USD',
      language: s.language ?? 'en',
      email: s.email ?? '',
      display_name: s.display_name ?? '',
      role: s.role ?? 'user',
      is_only_admin: s.is_only_admin ?? false,
    })
    if (isAdminEditing && userId) {
      settingsApi
        .getForUser(userId)
        .then((s) => setSettings(normalize(s)))
        .catch(() => setError(t('settings.loadFailed')))
        .finally(() => setLoading(false))
    } else {
      settingsApi
        .get()
        .then((s) => setSettings(normalize(s)))
        .catch(() => setError(t('settings.loadFailed')))
        .finally(() => setLoading(false))
    }
  }, [userId, isAdminEditing])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!settings) return
    setSaving(true)
    setError('')
    setMessage('')
    try {
      if (isAdminEditing && userId) {
        await settingsApi.updateForUser(userId, settings)
      } else {
        await settingsApi.update(settings)
        await refreshUser()
      }
      setMessage('settings.saved')
    } catch {
      setError(t('settings.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  useEffect(() => {
    if (!message) return
    const id = window.setTimeout(() => setMessage(''), 5000)
    return () => window.clearTimeout(id)
  }, [message])

  useEffect(() => {
    if (!passwordMessage) return
    const id = window.setTimeout(() => setPasswordMessage(''), 5000)
    return () => window.clearTimeout(id)
  }, [passwordMessage])

  async function handleChangePassword(e: React.FormEvent) {
    e.preventDefault()
    setPasswordError('')
    setPasswordMessage('')
    if (passwordNew !== passwordConfirm) {
      setPasswordError(t('settings.passwordsDoNotMatch'))
      return
    }
    if (passwordNew.length < 8) {
      setPasswordError(t('error.password_too_short'))
      return
    }
    setPasswordChanging(true)
    try {
      await apiChangePassword(passwordCurrent, passwordNew)
      setPasswordCurrent('')
      setPasswordNew('')
      setPasswordConfirm('')
      setPasswordMessage(t('settings.passwordChanged'))
    } catch (err) {
      const msg = err instanceof Error ? err.message : ''
      setPasswordError(msg.startsWith('error.') ? t(msg) : msg || t('settings.passwordChangeFailed'))
    } finally {
      setPasswordChanging(false)
    }
  }

  if (loading) return <div className="page"><p>{t('common.loading')}</p></div>
  if (error) return <div className="page"><p className="error">{error}</p></div>
  if (!settings) return null

  return (
    <div className="page">
      <header className="page-header">
        <h1>
          <Icon icon="mdi:cog" width={28} height={28} style={{ verticalAlign: 'middle', marginRight: 8 }} />
          {isAdminEditing && (settings.display_name !== undefined || settings.email !== undefined)
            ? t('settings.forUserDisplay', { displayName: settings.display_name ?? '', email: settings.email ?? '' })
            : isAdminEditing
              ? t('settings.forUser.empty')
              : t('settings.title')}
        </h1>
      </header>
      <p className="muted">
        {isAdminEditing ? t('settings.weightUnitHelpAdmin') : t('settings.weightUnitHelp')}
      </p>
      <form onSubmit={handleSubmit} className="form" style={{ maxWidth: 360 }}>
        {isAdminEditing && (
          <>
            <label>
              Email
              <input
                type="email"
                value={settings.email ?? ''}
                onChange={(e) => setSettings((s) => s && { ...s, email: e.target.value })}
                className="input"
                placeholder="user@example.com"
              />
            </label>
            <label>
              Role
              <select
                value={settings.role ?? 'user'}
                onChange={(e) => setSettings((s) => s && { ...s, role: e.target.value })}
                disabled={settings.is_only_admin}
                title={settings.is_only_admin ? 'Cannot remove the only admin' : undefined}
              >
                <option value="user">user</option>
                <option value="admin">admin</option>
              </select>
              {settings.is_only_admin && (
                <span className="muted" style={{ fontSize: '0.875rem', display: 'block', marginTop: 4 }}>
                  Cannot change role; this is the only admin.
                </span>
              )}
            </label>
          </>
        )}
        <label>
          {t('settings.weightUnit')}
          <select
            value={settings.weight_unit}
            onChange={(e) => setSettings((s) => s && { ...s, weight_unit: e.target.value as WeightUnit })}
          >
            <option value="lbs">{t('settings.pounds')}</option>
            <option value="kg">{t('settings.kilograms')}</option>
          </select>
        </label>
        <label>
          {t('settings.currency')}
          <select
            value={settings.currency || 'USD'}
            onChange={(e) => setSettings((s) => s && { ...s, currency: e.target.value })}
          >
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
            <option value="GBP">GBP</option>
            <option value="CAD">CAD</option>
            <option value="AUD">AUD</option>
            <option value="JPY">JPY</option>
          </select>
        </label>
        <label>
          {t('settings.language')}
          <select
            value={settings.language || 'en'}
            onChange={(e) => setSettings((s) => s && { ...s, language: e.target.value })}
          >
            <option value="en">English</option>
            <option value="es">Español</option>
            <option value="fr">Français</option>
            <option value="de">Deutsch</option>
          </select>
        </label>
        {message && <p className="message" style={{ color: 'var(--dark-accent)' }}>{t(message)}</p>}
        <div className="form-actions">
          <button type="submit" className="btn btn-primary" disabled={saving}>
            {saving ? t('common.saving') : t('common.save')}
          </button>
        </div>
      </form>

      {!isAdminEditing && (
        <section className="section" style={{ maxWidth: 360, marginTop: '1.5rem' }}>
          <h2 style={{ marginTop: 0 }}>{t('settings.changePassword')}</h2>
          <form onSubmit={handleChangePassword} className="form">
            <label>
              {t('settings.currentPassword')}
              <input
                type="password"
                value={passwordCurrent}
                onChange={(e) => setPasswordCurrent(e.target.value)}
                autoComplete="current-password"
                required
              />
            </label>
            <label>
              {t('settings.newPassword')}
              <input
                type="password"
                value={passwordNew}
                onChange={(e) => setPasswordNew(e.target.value)}
                autoComplete="new-password"
                minLength={8}
                required
              />
            </label>
            <label>
              {t('settings.confirmPassword')}
              <input
                type="password"
                value={passwordConfirm}
                onChange={(e) => setPasswordConfirm(e.target.value)}
                autoComplete="new-password"
                minLength={8}
                required
              />
            </label>
            {passwordError && <p className="error">{passwordError}</p>}
            {passwordMessage && <p className="message" style={{ color: 'var(--dark-accent)' }}>{passwordMessage}</p>}
            <div className="form-actions">
              <button type="submit" className="btn btn-primary" disabled={passwordChanging}>
                {passwordChanging ? t('common.saving') : t('settings.changePassword')}
              </button>
            </div>
          </form>
        </section>
      )}

      {pwa && !pwa.isStandalone && (
        <section className="section" style={{ maxWidth: 360, marginTop: '1.5rem' }}>
          <h2 style={{ marginTop: 0 }}>Install app</h2>
          <p className="muted">Add Pet Medical to your home screen for quick access.</p>
          {pwa.deferredPrompt ? (
            <button type="button" className="btn btn-primary" onClick={async () => { await pwa.deferredPrompt!.prompt(); pwa.setDismissed(true) }}>
              Install Pet Medical
            </button>
          ) : (
            <>
              {pwa.isIOS ? (
                <p className="muted" style={{ fontSize: '0.9rem' }}>
                  In Safari: tap the Share button, then &quot;Add to Home Screen&quot;.
                </p>
              ) : pwa.isMobile ? (
                <p className="muted" style={{ fontSize: '0.9rem' }}>
                  In Chrome: tap the three dots (⋮) → look for &quot;Add to Home screen&quot; or &quot;Install app&quot;. If you don&apos;t see it, use the site over HTTPS and try again—Chrome shows the option when the app is installable.
                </p>
              ) : null}
              <button type="button" className="btn btn-secondary btn-sm" onClick={pwa.clearDismissed}>
                Show install banner again
              </button>
            </>
          )}
        </section>
      )}
    </div>
  )
}
