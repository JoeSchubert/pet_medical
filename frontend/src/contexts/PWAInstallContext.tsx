import { createContext, useCallback, useContext, useEffect, useState } from 'react'

const DISMISS_KEY = 'pwa-install-dismissed'

export function isStandalone(): boolean {
  if (typeof window === 'undefined') return true
  return (
    window.matchMedia('(display-mode: standalone)').matches ||
    (navigator as { standalone?: boolean }).standalone === true
  )
}

export function isIOS(): boolean {
  if (typeof navigator === 'undefined') return false
  return /iPad|iPhone|iPod/.test(navigator.userAgent) || (navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1)
}

export function isMobile(): boolean {
  if (typeof navigator === 'undefined') return false
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent) || (navigator.maxTouchPoints ?? 0) > 0
}

type BeforeInstallPromptEvent = Event & { prompt: () => Promise<{ outcome: string }> }

type PWAInstallContextValue = {
  deferredPrompt: BeforeInstallPromptEvent | null
  dismissed: boolean
  clearDismissed: () => void
  setDismissed: (v: boolean) => void
  isStandalone: boolean
  isMobile: boolean
  isIOS: boolean
}

const PWAInstallContext = createContext<PWAInstallContextValue | null>(null)

export function PWAInstallProvider({ children }: { children: React.ReactNode }) {
  const [deferredPrompt, setDeferredPromptState] = useState<BeforeInstallPromptEvent | null>(null)
  const [dismissed, setDismissedState] = useState(() => {
    if (typeof localStorage === 'undefined') return true
    return localStorage.getItem(DISMISS_KEY) === '1'
  })
  const [standalone, setStandalone] = useState(true)
  const [mobile, setMobile] = useState(false)
  const [ios, setIos] = useState(false)

  const setDismissed = useCallback((v: boolean) => {
    setDismissedState(v)
    if (typeof localStorage !== 'undefined') {
      if (v) localStorage.setItem(DISMISS_KEY, '1')
      else localStorage.removeItem(DISMISS_KEY)
    }
  }, [])

  const clearDismissed = useCallback(() => {
    setDismissedState(false)
    if (typeof localStorage !== 'undefined') localStorage.removeItem(DISMISS_KEY)
  }, [])

  useEffect(() => {
    setStandalone(isStandalone())
    setMobile(isMobile())
    setIos(isIOS())
  }, [])

  useEffect(() => {
    const handler = (e: Event) => {
      e.preventDefault()
      setDeferredPromptState(e as BeforeInstallPromptEvent)
    }
    window.addEventListener('beforeinstallprompt', handler)
    return () => window.removeEventListener('beforeinstallprompt', handler)
  }, [])

  const value: PWAInstallContextValue = {
    deferredPrompt,
    dismissed,
    clearDismissed,
    setDismissed,
    isStandalone: standalone,
    isMobile: mobile,
    isIOS: ios,
  }

  return (
    <PWAInstallContext.Provider value={value}>
      {children}
    </PWAInstallContext.Provider>
  )
}

export function usePWAInstall() {
  const ctx = useContext(PWAInstallContext)
  return ctx
}
