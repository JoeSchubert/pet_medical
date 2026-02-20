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

  if (loading) return <div className="page"><p role="status">{t('common.loading')}</p></div>
  if (error) return <div className="page"><p className="error" role="alert">{error}</p></div>
  if (!settings) return null

  return (
    <div className="page" aria-label="Settings">
      <header className="page-header">
        <h1 id="settings-title">
          <Icon icon="mdi:cog" width={28} height={28} style={{ verticalAlign: 'middle', marginRight: 8 }} aria-hidden />
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
      <div className="settings-cards">
        <div className="settings-card">
          <form onSubmit={handleSubmit} className="form" aria-labelledby="settings-title">
        {isAdminEditing && (
          <>
            <label htmlFor="settings-email">
              Email
              <input
                id="settings-email"
                type="email"
                value={settings.email ?? ''}
                onChange={(e) => setSettings((s) => s && { ...s, email: e.target.value })}
                className="input"
                placeholder="user@example.com"
              />
            </label>
            <label htmlFor="settings-role">
              Role
              <select
                id="settings-role"
                value={settings.role ?? 'user'}
                onChange={(e) => setSettings((s) => s && { ...s, role: e.target.value })}
                disabled={settings.is_only_admin}
                title={settings.is_only_admin ? 'Cannot remove the only admin' : undefined}
                aria-describedby={settings.is_only_admin ? 'settings-role-desc' : undefined}
              >
                <option value="user">user</option>
                <option value="admin">admin</option>
              </select>
              {settings.is_only_admin && (
                <span id="settings-role-desc" className="muted" style={{ fontSize: '0.875rem', display: 'block', marginTop: 4 }}>
                  Cannot change role; this is the only admin.
                </span>
              )}
            </label>
          </>
        )}
        <label htmlFor="settings-weight-unit">
          {t('settings.weightUnit')}
          <select
            id="settings-weight-unit"
            value={settings.weight_unit}
            onChange={(e) => setSettings((s) => s && { ...s, weight_unit: e.target.value as WeightUnit })}
          >
            <option value="lbs">{t('settings.pounds')}</option>
            <option value="kg">{t('settings.kilograms')}</option>
          </select>
        </label>
        <label htmlFor="settings-currency">
          {t('settings.currency')}
          <select
            id="settings-currency"
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
        <label htmlFor="settings-language">
          {t('settings.language')}
          <select
            id="settings-language"
            value={settings.language || 'en'}
            onChange={(e) => setSettings((s) => s && { ...s, language: e.target.value })}
          >
            <option value="en">English</option>
            <option value="es">Español</option>
            <option value="fr">Français</option>
            <option value="de">Deutsch</option>
          </select>
        </label>
        {message && <p className="message" style={{ color: 'var(--dark-accent)' }} role="status">{t(message)}</p>}
        <div className="form-actions">
          <button type="submit" className="btn btn-primary" disabled={saving} aria-busy={saving}>
            {saving ? t('common.saving') : t('common.save')}
          </button>
        </div>
          </form>
        </div>

        {!isAdminEditing && (
        <section className="section settings-card" aria-labelledby="settings-password-heading">
          <h2 id="settings-password-heading" style={{ marginTop: 0 }}>{t('settings.changePassword')}</h2>
          <form onSubmit={handleChangePassword} className="form">
            <label htmlFor="settings-current-password">
              {t('settings.currentPassword')}
              <input
                id="settings-current-password"
                type="password"
                value={passwordCurrent}
                onChange={(e) => setPasswordCurrent(e.target.value)}
                autoComplete="current-password"
                required
              />
            </label>
            <label htmlFor="settings-new-password">
              {t('settings.newPassword')}
              <input
                id="settings-new-password"
                type="password"
                value={passwordNew}
                onChange={(e) => setPasswordNew(e.target.value)}
                autoComplete="new-password"
                minLength={8}
                required
                aria-describedby={passwordError ? 'settings-password-error' : undefined}
              />
            </label>
            <label htmlFor="settings-confirm-password">
              {t('settings.confirmPassword')}
              <input
                id="settings-confirm-password"
                type="password"
                value={passwordConfirm}
                onChange={(e) => setPasswordConfirm(e.target.value)}
                autoComplete="new-password"
                minLength={8}
                required
              />
            </label>
            {passwordError && <p id="settings-password-error" className="error" role="alert">{passwordError}</p>}
            {passwordMessage && <p className="message" style={{ color: 'var(--dark-accent)' }} role="status">{t(passwordMessage)}</p>}
            <div className="form-actions">
              <button type="submit" className="btn btn-primary" disabled={passwordChanging} aria-busy={passwordChanging}>
                {passwordChanging ? t('common.saving') : t('settings.changePassword')}
              </button>
            </div>
          </form>
        </section>
        )}

        {pwa && !pwa.isStandalone && (
        <section className="section settings-card" aria-labelledby="settings-install-heading">
          <h2 id="settings-install-heading" style={{ marginTop: 0 }}>Install app</h2>
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
    </div>
  )
}
