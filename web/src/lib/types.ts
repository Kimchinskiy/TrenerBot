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

export interface ScheduleEntry {
  id: number
  date: string
  time: string
  client_id: number
  client_name: string
  coach_id?: number | null
  duration: number
  status: string
  group_id?: number | null
  group_name?: string | null
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
  user_id?: number
  client?: Client
  children?: Client[]
  telegram_id?: string | null
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

export type Role = MeResult['role']

export interface Recipient {
  client_id: number
  full_name: string
  user_id: number | null
}

export interface NotificationPreview {
  total: number
  recipients: Recipient[]
}

export interface SendResult {
  total: number
  enqueued: number
  skipped: number
  errors: number
}

export interface DateAttendanceClient {
  client_id: number
  full_name: string
  time: string
  present: boolean | null
  photo?: string | null
}

export interface DateAttendanceResponse {
  date: string
  clients: DateAttendanceClient[]
}

export interface SaveAttendanceEntry {
  client_id: number
  present: boolean
}

export interface CoachSubscription {
  id: number
  coach_id: number
  status: 'trial' | 'active' | 'expired' | 'canceled'
  trial_start: string
  trial_end?: string | null
  paid_until?: string | null
  created_at: string
}

export interface CoachOnboarding {
  is_coach: boolean
  message?: string
  coach?: { id: number; full_name: string }
  subscription?: CoachSubscription
  active?: boolean
  days_left?: number
}

export interface ChildLessonStatus {
  client_id: number
  full_name: string
  date: string
  time: string
  duration: number
  status: string
  is_today: boolean
  is_ongoing: boolean
  minutes_left?: number | null
  minutes_until?: number | null
  has_lesson_today: boolean
}

export interface SocialLink {
  platform: string
  url?: string | null
  enabled: boolean
}

export interface ParentNotifPref {
  id: number
  parent_user_id: number
  child_id: number
  lesson_start: boolean
  lesson_end_15: boolean
  lesson_missed: boolean
}

export interface Group {
  id: number
  name?: string | null
  coach_id?: number | null
  max_members?: number | null
  schedule?: string | null
  price?: number | null
  location?: string | null
  active: number
  member_count: number
}

export interface ClientSubscription {
  id: number
  client_id: number
  type: string
  price: number
  bought_at: string
  ends_at: string
  lessons_left: number
  freeze: number
  created_at: string
}

export interface StatisticsResponse {
  period: 'week' | 'month' | 'year'
  date_from: string
  date_to: string
  trainings: MetricValue
  clients: MetricValue
  income: IncomeMetric
  attendance: MetricValue
  debtors: DebtorsSummary
  income_chart: ChartPoint[]
  quick_overview: QuickOverview
}

export interface MetricValue {
  value: number
  change: number
  label: string
}

export interface IncomeMetric {
  value: number
  change: number
  label: string
}

export interface DebtorsSummary {
  count: number
  total_debt: number
  items: DebtorItem[]
}

export interface DebtorItem {
  client_id: number
  full_name: string
  phone?: string
  debt: number
  ends_at?: string
}

export interface ChartPoint {
  label: string
  value: number
}

export interface QuickOverview {
  new_clients: number
  average_check: number
  canceled_count: number
  avg_attendance: number
  avg_group_size: number
  busiest_day: string
  popular_time: string
}

export interface Lead {
  id: number
  telegram_id: string
  full_name: string
  phone?: string | null
  target_name?: string | null
  target_age?: number | null
  target_level: string
  reg_type: string
  status: string
  created_at: string
}

export interface GroupMember {
  id: number
  group_id: number
  client_id: number
  role: string
  joined_at: string
  client_name?: string | null
}
