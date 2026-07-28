import type {
  Lesson,
  MeResult,
  Attendance,
  Debtor,
  WaitingItem,
  WellbeingEntry,
  Client,
  AuthTokens,
  ScheduleEntry,
  Recipient,
  NotificationPreview,
  SendResult,
  DateAttendanceResponse,
  SaveAttendanceEntry,
  CoachOnboarding,
  CoachSubscription,
  ChildLessonStatus,
  ParentNotifPref,
  SocialLink,
  Group,
  GroupMember,
  ClientSubscription,
  StatisticsResponse,
  Lead,
} from './types'

const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api'

const TOKEN_KEY = 'trenerbot_token'
const REFRESH_KEY = 'trenerbot_refresh'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY)
}

export function setRefreshToken(token: string) {
  localStorage.setItem(REFRESH_KEY, token)
}

export function clearRefreshToken() {
  localStorage.removeItem(REFRESH_KEY)
}

// Persist the token pair returned by any auth endpoint.
export function storeTokens(t: { access_token: string; refresh_token: string }) {
  setToken(t.access_token)
  setRefreshToken(t.refresh_token)
}

export function clearSession() {
  clearToken()
  clearRefreshToken()
}

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

// Attempt to refresh the access token using the stored refresh token.
// Returns the new access token or null when refresh is impossible.
let refreshInFlight: Promise<string | null> | null = null

async function refreshAccessToken(): Promise<string | null> {
  const refresh = getRefreshToken()
  if (!refresh) return null
  if (refreshInFlight) return refreshInFlight
  refreshInFlight = (async () => {
    try {
      const res = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refresh }),
      })
      if (!res.ok) {
        clearSession()
        return null
      }
      const data = (await res.json()) as { access_token: string; refresh_token: string }
      storeTokens(data)
      return data.access_token
    } catch {
      return null
    } finally {
      refreshInFlight = null
    }
  })()
  return refreshInFlight
}

async function request<T>(path: string, options: RequestInit = {}, retry = true): Promise<T> {
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
    // Try a one-time transparent refresh before failing.
    if (retry) {
      const newToken = await refreshAccessToken()
      if (newToken) {
        return request<T>(path, options, false)
      }
    }
    clearSession()
    // Notify the auth state listener (e.g. redirect to login).
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('auth:unauthorized'))
    }
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

  // --- Notifications (coach broadcast) ---

  notificationsPreview(body: { filter: string; group_id?: number; client_ids?: number[] }) {
    return request<NotificationPreview>('/notifications/preview', {
      method: 'POST',
      body: JSON.stringify(body),
    })
  },
  notificationsSend(body: { filter: string; group_id?: number; client_ids?: number[]; title: string; text: string }) {
    return request<SendResult>('/notifications/send', {
      method: 'POST',
      body: JSON.stringify(body),
    })
  },

  // Website auth (primary): phone + password.
  register(phone: string, password: string, firstName: string, lastName: string) {
    return request<AuthTokens>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ phone, password, first_name: firstName, last_name: lastName }),
    })
  },
  login(phone: string, password: string) {
    return request<AuthTokens>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ phone, password }),
    })
  },
  logout(refreshToken: string) {
    return request<{ status: string }>('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
  },

  // Telegram Login Widget (website login, not Mini App). `fields` is the widget payload.
  loginTelegramWidget(fields: Record<string, string>) {
    return request<AuthTokens>('/auth/telegram-widget', {
      method: 'POST',
      body: JSON.stringify(fields),
    })
  },

  // Additional method: exchange Telegram Mini App initData for a token pair.
  loginWebApp(initData: string) {
    return request<AuthTokens>('/auth/telegram-webapp', {
      method: 'POST',
      body: JSON.stringify({ init_data: initData }),
    })
  },
}

export const endpoints = {
  me: () => request<MeResult>('/clients/me'),
  lessons: (from: string, to: string) => request<Lesson[]>(`/lessons?from=${from}&to=${to}`),
  schedule: (from: string, to: string) => request<ScheduleEntry[]>(`/schedule?from=${from}&to=${to}`),
  notificationsPreview: (body: { filter: string; group_id?: number; client_ids?: number[] }) =>
    request<NotificationPreview>('/notifications/preview', {
      method: 'POST',
      body: JSON.stringify(body),
    }),
  notificationsSend: (body: { filter: string; group_id?: number; client_ids?: number[]; title: string; text: string }) =>
    request<SendResult>('/notifications/send', {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  // Coach
  coachOnboarding: () => request<CoachOnboarding>('/coach/onboarding'),
  upgradeToCoach: (fullName: string, sport: string) =>
    request<{ status: string; role: string; coach_id: number }>('/coach/upgrade', {
      method: 'POST', body: JSON.stringify({ full_name: fullName, sport }),
    }),
  coachSubscription: () => request<{ subscription: CoachSubscription; active: boolean }>('/coach/subscription'),
  startCoachTrial: () => request<{ status: string; subscription: CoachSubscription }>('/coach/subscription/trial', { method: 'POST' }),
  activateCoachSubscription: (days: number) =>
    request<{ status: string; subscription: CoachSubscription }>('/coach/subscription/activate', {
      method: 'POST', body: JSON.stringify({ days }),
    }),

  // Parent
  upgradeToParent: () => request<{ status: string; role: string }>('/parent/upgrade', { method: 'POST' }),
  linkChild: (fullName: string, birthDate: string) =>
    request<{ status: string; child_id: number; child_name: string }>('/parent/link', {
      method: 'POST', body: JSON.stringify({ full_name: fullName, birth_date: birthDate }),
    }),
  childrenLessonStatus: () => request<ChildLessonStatus[]>('/parent/children/status'),
  parentNotifPrefs: () => request<ParentNotifPref[]>('/parent/notif-prefs'),
  saveParentNotifPref: (pref: { child_id: number; lesson_start?: boolean; lesson_end_15?: boolean; lesson_missed?: boolean }) =>
    request<{ status: string }>('/parent/notif-prefs', {
      method: 'POST', body: JSON.stringify(pref),
    }),
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
  createScheduleEntry: (data: { client_id?: number; group_id?: number; date: string; time: string; duration?: number }) =>
    request<{ ids: number[] }>('/schedule', { method: 'POST', body: JSON.stringify(data) }),
  socialMedia: () => request<Record<string, string>>('/social-media'),
  saveSocialLinks: (links: SocialLink[]) =>
    request<{ status: string }>('/social-media', {
      method: 'POST', body: JSON.stringify({ links }),
    }),
  clients: () => request<Client[]>('/clients'),
  groups: () => request<Group[]>('/groups'),
  group: (id: number) => request<Group>(`/groups/${id}`),
  createGroup: (data: { name: string; coach_id?: number; max_members?: number; schedule?: string; price?: number; location?: string; active?: number }) =>
    request<{ id: number }>('/groups', { method: 'POST', body: JSON.stringify(data) }),
  updateGroup: (id: number, data: { name?: string; coach_id?: number; max_members?: number; schedule?: string; price?: number; location?: string; active?: number }) =>
    request<{ status: string }>(`/groups/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteGroup: (id: number) => request<{ status: string }>(`/groups/${id}`, { method: 'DELETE' }),
  groupClients: (id: number) => request<GroupMember[]>(`/groups/${id}/clients`),
  groupAvailableClients: (id: number) => request<Client[]>(`/groups/${id}/available-clients`),
  addClientToGroup: (id: number, clientId: number, role?: string) =>
    request<{ status: string }>(`/groups/${id}/clients`, { method: 'POST', body: JSON.stringify({ client_id: clientId, role }) }),
  removeClientFromGroup: (id: number, clientId: number) =>
    request<{ status: string }>(`/groups/${id}/clients`, { method: 'DELETE', body: JSON.stringify({ client_id: clientId }) }),
  clientSubscriptions: (clientId: number) => request<ClientSubscription[]>(`/clients/${clientId}/subscriptions`),
  createClientSubscription: (data: { client_id: number; type: string; price?: number; ends_at?: string; lessons_left?: number }) =>
    request<{ id: number }>(`/clients/${data.client_id}/subscriptions`, { method: 'POST', body: JSON.stringify(data) }),
  updateClientSubscription: (data: { id: number; client_id: number; type?: string; price?: number; ends_at?: string; lessons_left?: number; freeze?: number }) =>
    request<{ status: string }>(`/clients/${data.client_id}/subscriptions`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteClientSubscription: (clientId: number, id: number) =>
    request<{ status: string }>(`/clients/${clientId}/subscriptions/${id}`, { method: 'DELETE' }),
  leads: () => request<Lead[]>('/leads'),
  approveLead: (id: number) => request<Lead>(`/leads/${id}`, { method: 'POST', body: JSON.stringify({ action: 'approve' }) }),
  rejectLead: (id: number) => request<{ status: string }>(`/leads/${id}`, { method: 'POST', body: JSON.stringify({ action: 'reject' }) }),
  statistics: (period: string) => request<StatisticsResponse>(`/statistics?period=${period}`),
  faq: (q: string) => request<{ answer: string }>(`/faq?q=${encodeURIComponent(q)}`),
  dateAttendance: (date: string) => request<DateAttendanceResponse>(`/attendance/date/${date}`),
  saveDateAttendance: (date: string, entries: SaveAttendanceEntry[]) =>
    request<{ status: string }>('/attendance/date', {
      method: 'POST',
      body: JSON.stringify({ date, entries }),
    }),
}

