import { api, setToken, getToken } from './api'
import type { AuthResult } from './types'

// Telegram WebApp SDK (lazy access so the app still loads in a normal browser).
function webApp(): any {
  return (window as any).Telegram?.WebApp
}

export function getInitData(): string {
  const tw = webApp()
  return tw?.initData ?? ''
}

export interface AuthState {
  token: string
  role: string
}

// Log in via the Mini App initData. Returns the resolved auth state, or null
// if we are not running inside Telegram and have no stored token.
export async function login(): Promise<AuthState> {
  const existing = getToken()
  if (existing) {
    return { token: existing, role: '' }
  }
  const initData = getInitData()
  if (!initData) {
    throw new Error('Не удалось получить данные Telegram. Откройте приложение через бота.')
  }
  const res = await api.loginWebApp(initData)
  setToken(res.token)
  return { token: res.token, role: res.user.role }
}

export function logout() {
  const tw = webApp()
  tw?.MainButton?.hide?.()
  localStorage.removeItem('trenerbot_token')
}
