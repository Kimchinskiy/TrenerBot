import { api, storeTokens, clearSession, getToken, getRefreshToken } from './api'
import type { AuthTokens } from './types'

// Telegram WebApp SDK (lazy access so the app still works in a normal browser).
function webApp(): any {
  return (window as any).Telegram?.WebApp
}

export function getInitData(): string {
  const tw = webApp()
  return tw?.initData ?? ''
}

// isInsideTelegram reports whether the app is running inside the Telegram Mini App.
export function isInsideTelegram(): boolean {
  return !!getInitData()
}

// tryAutoLogin attempts a session without user interaction:
//   1. a previously stored access token (returning visitor);
//   2. Telegram Mini App initData (additional method when opened from Telegram).
// Returns true if a session is active, false if the user must sign in on the site.
export async function tryAutoLogin(): Promise<boolean> {
  if (getToken()) return true

  const initData = getInitData()
  if (initData) {
    try {
      const res = await api.loginWebApp(initData)
      storeTokens(res)
      return true
    } catch {
      // fall through to interactive login
    }
  }
  return false
}

// Primary website auth: phone + password.
export async function loginWithPassword(phone: string, password: string): Promise<AuthTokens> {
  const res = await api.login(phone, password)
  storeTokens(res)
  return res
}

export async function registerWithPassword(
  phone: string,
  password: string,
  firstName: string,
  lastName: string,
): Promise<AuthTokens> {
  const res = await api.register(phone, password, firstName, lastName)
  storeTokens(res)
  return res
}

// Telegram Login Widget (website login, not Mini App).
export async function loginWithTelegramWidget(fields: Record<string, string>): Promise<AuthTokens> {
  const res = await api.loginTelegramWidget(fields)
  storeTokens(res)
  return res
}

export async function logout() {
  webApp()?.MainButton?.hide?.()
  const refresh = getRefreshToken()
  if (refresh) {
    // Revoke the refresh token server-side; ignore network errors.
    try {
      await api.logout(refresh)
    } catch {
      /* ignore */
    }
  }
  clearSession()
}
