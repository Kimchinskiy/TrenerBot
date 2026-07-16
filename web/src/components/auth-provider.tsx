'use client'

import * as React from 'react'
import { getToken, clearSession } from '@/lib/api'
import { tryAutoLogin } from '@/lib/auth'
import { useMe } from '@/lib/hooks'
import type { Role } from '@/lib/types'

type AuthStatus = 'loading' | 'authed' | 'guest'

interface AuthContextValue {
  status: AuthStatus
  role: Role | null
  isAuthed: boolean
  /** Mark the session as authenticated after a successful login/register. */
  setAuthed: () => void
  /** Force a guest state and clear tokens. */
  forceGuest: () => void
}

const AuthContext = React.createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = React.useState<AuthStatus>('loading')
  const me = useMe()
  const role = me.data?.role ?? null

  React.useEffect(() => {
    tryAutoLogin()
      .then((ok) => setStatus(ok ? 'authed' : 'guest'))
      .catch(() => setStatus('guest'))
  }, [])

  React.useEffect(() => {
    const onUnauthorized = () => {
      clearSession()
      setStatus('guest')
    }
    window.addEventListener('auth:unauthorized', onUnauthorized)
    return () => window.removeEventListener('auth:unauthorized', onUnauthorized)
  }, [])

  const value: AuthContextValue = {
    status,
    role,
    isAuthed: status === 'authed',
    setAuthed: () => setStatus('authed'),
    forceGuest: () => {
      clearSession()
      setStatus('guest')
    },
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = React.useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
