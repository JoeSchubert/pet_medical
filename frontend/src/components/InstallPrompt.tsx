import { useEffect, useState } from 'react'
import { Icon } from '@iconify/react'
import { usePWAInstall } from '../contexts/PWAInstallContext'

export default function InstallPrompt() {
  const ctx = usePWAInstall()
  const [showIOSHint, setShowIOSHint] = useState(false)

  useEffect(() => {
    if (!ctx || ctx.isStandalone || ctx.dismissed) return
    if (ctx.isIOS) {
      const isSafari = /Safari/.test(navigator.userAgent) && !/Chrome/.test(navigator.userAgent)
      if (isSafari) setShowIOSHint(true)
    }
  }, [ctx?.isStandalone, ctx?.dismissed, ctx?.isIOS])

  if (!ctx || ctx.isStandalone || ctx.dismissed) return null
  if (!ctx.isMobile) return null

  const showBar = ctx.isMobile || ctx.deferredPrompt
  if (!showBar && !showIOSHint) return null

  const handleInstall = async () => {
    if (!ctx.deferredPrompt) return
    await ctx.deferredPrompt.prompt()
    ctx.setDismissed(true)
  }

  const handleDismiss = () => {
    ctx.setDismissed(true)
    setShowIOSHint(false)
  }

  return (
    <div className="install-prompt" role="banner">
      <div className="install-prompt-inner">
        <span className="install-prompt-logo">🐾</span>
        <div className="install-prompt-text">
          {ctx.deferredPrompt ? (
            <>
              <strong>Install Pet Medical</strong>
              <span>Use the app from your home screen for quick access.</span>
            </>
          ) : showIOSHint ? (
            <>
              <strong>Add to Home Screen</strong>
              <span>Tap Share in Safari, then &quot;Add to Home Screen&quot;</span>
            </>
          ) : ctx.isMobile ? (
            <>
              <strong>Add to Home Screen</strong>
              <span>Android: menu (⋮) → &quot;Add to Home screen&quot; or &quot;Install app&quot;. iOS: Safari → Share → Add to Home Screen.</span>
            </>
          ) : null}
        </div>
        <div className="install-prompt-actions">
          {ctx.deferredPrompt && (
            <button type="button" className="btn btn-primary btn-sm" onClick={handleInstall}>
              Install
            </button>
          )}
          <button
            type="button"
            className="install-prompt-dismiss"
            onClick={handleDismiss}
            aria-label="Dismiss"
          >
            <Icon icon="mdi:close" width={20} height={20} />
          </button>
        </div>
      </div>
    </div>
  )
}
