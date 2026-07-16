export interface Lesson {
  id: number
  date: string
  time: string
  coach_id?: number | null
  duration: number
  status: string
  location?: string | null
  comment?: string | null
  group_id?: number | null
}

export interface Client {
  id: number
  user_id?: number | null
  full_name: string
  age?: number | null
  phone?: string | null
  medical_limits?: string | null
  status: string
  source?: string | null
  bot_access: boolean
  subscription_ends_at?: string | null
}

export interface MeResult {
  role: 'admin' | 'coach' | 'client' | 'parent'
  client?: Client
  children?: Client[]
}

export interface Attendance {
  lesson_id: number
  client_id: number
  present: boolean
  marked_at?: string | null
  marked_by?: number | null
}

export interface Debtor {
  client_id: number
  name: string
  phone: string
  status: string
  missed_count: number
  missed_dates: string
}

export interface WaitingItem {
  id: number
  client_id: number
  group_id?: number | null
  position: number
  created_at: string
  name: string
  phone: string
}

export interface WellbeingEntry {
  id: number
  lesson_id: number
  note: string
  created_at: string
  date: string
  time: string
}

export interface User {
  id: number
  phone?: string | null
  telegram_id?: string | null
  max_id?: string | null
  first_name?: string | null
  last_name?: string | null
  avatar_url?: string | null
  role: Role
  created_at?: string
  updated_at?: string | null
}

export interface AuthTokens {
  access_token: string
  refresh_token: string
  user: User
}

export interface AuthResult {
  user: { id: number; role: string }
  client: Client | null
  token: string
}

export type Role = MeResult['role']
