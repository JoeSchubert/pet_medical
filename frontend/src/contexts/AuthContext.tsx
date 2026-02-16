import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react'
import * as api from '../api/client'
import { logAuth } from '../lib/log'

interface AuthContextValue {
  user: api.User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<api.User | null>(null)
  const [loading, setLoading] = useState(true)
  const refreshVersion = useRef(0)

  const refreshUser = useCallback(async () => {
    const version = ++refreshVersion.current
    logAuth('refreshUser started version=', version)
    try {
      const u = await api.getMe()
      logAuth('refreshUser getMe success version=', version)
      if (version === refreshVersion.current) setUser(u)
    } catch {
      logAuth('refreshUser getMe failed, trying refresh')
      try {
        const data = await api.refresh()
        if (version === refreshVersion.current) setUser(data.user)
      } catch {
        if (version === refreshVersion.current) setUser(null)
      }
    } finally {
      if (version === refreshVersion.current) setLoading(false)
    }
  }, [])

  useEffect(() => {
    refreshUser()
  }, [refreshUser])

  const login = useCallback(async (email: string, password: string) => {
    logAuth('login() called email=', email)
    const res = await api.login(email, password)
    logAuth('login() api.login returned user=', res.user?.display_name, 'setting user, bumping version')
    setUser(res.user)
    refreshVersion.current += 1
    logAuth('login() done version now=', refreshVersion.current)
  }, [])

  const logout = useCallback(async () => {
    await api.logout()
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
