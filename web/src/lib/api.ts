import type {
  Lesson,
  MeResult,
  Attendance,
  Debtor,
  WaitingItem,
  WellbeingEntry,
  Client,
} from './types'

const API_BASE = import.meta.env.VITE_API_URL || '/api'

const TOKEN_KEY = 'trenerbot_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers })
  if (res.status === 401) {
    clearToken()
    throw new ApiError(401, 'unauthorized')
  }
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const data = await res.json()
      if (data?.error) msg = data.error
    } catch {
      /* ignore */
    }
    throw new ApiError(res.status, msg)
  }
  if (res.status === 204) return undefined as T
  const text = await res.text()
  return text ? (JSON.parse(text) as T) : (undefined as T)
}

export const api = {
  base: API_BASE,

  // Public: exchange Telegram initData for a JWT.
  loginWebApp(initData: string) {
    return request<{ user: { id: number; role: string }; client: MeResult['client']; token: string }>(
      '/auth/telegram-webapp',
      { method: 'POST', body: JSON.stringify({ init_data: initData }) },
    )
  },
}

export const endpoints = {
  me: () => request<MeResult>('/clients/me'),
  lessons: (from: string, to: string) => request<Lesson[]>(`/lessons?from=${from}&to=${to}`),
  attendance: (lessonId: number) => request<Attendance[]>(`/lessons/${lessonId}/attendance`),
  markAttendance: (lessonId: number, clientId: number, present: boolean) =>
    request<{ status: string }>(`/lessons/${lessonId}/attendance`, {
      method: 'POST',
      body: JSON.stringify({ client_id: clientId, present }),
    }),
  debtors: (days: number) => request<Debtor[]>(`/debtors?days=${days}`),
  waitingList: () => request<WaitingItem[]>('/waiting-list'),
  addWaitingList: (clientId: number) =>
    request<{ id: number }>('/waiting-list', { method: 'POST', body: JSON.stringify({ client_id: clientId }) }),
  removeWaitingList: (id: number) => request<{ status: string }>(`/waiting-list/${id}`, { method: 'DELETE' }),
  wellbeing: (lessonId: number, wellbeing: number, note: string) =>
    request<{ status: string }>('/wellbeing', {
      method: 'POST',
      body: JSON.stringify({ lesson_id: lessonId, wellbeing, note }),
    }),
  wellbeingHistory: (clientId: number) => request<WellbeingEntry[]>(`/wellbeing/${clientId}`),
  messageCoach: (from: string, text: string) =>
    request<{ status: string }>('/messages/coach', {
      method: 'POST',
      body: JSON.stringify({ from, text }),
    }),
  socialMedia: () => request<Record<string, string>>('/social-media'),
  clients: () => request<Client[]>('/clients'),
  faq: (q: string) => request<{ answer: string }>(`/faq?q=${encodeURIComponent(q)}`),
}
