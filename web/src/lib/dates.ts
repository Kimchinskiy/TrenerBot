export function isoDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function weekRange(base = new Date()): { from: string; to: string } {
  const from = new Date(base)
  const day = from.getDay() === 0 ? 7 : from.getDay()
  from.setDate(from.getDate() - day + 1)
  const to = new Date(from)
  to.setDate(to.getDate() + 7)
  return { from: isoDate(from), to: isoDate(to) }
}

export const WEEKDAYS_RU = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']
export const WEEKDAYS_RU_FULL = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота']
export const MONTHS_RU = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня', 'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря']
export const MONTHS_RU_NOM = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь', 'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь']

export function prettyDate(iso: string): string {
  const [y, m, d] = iso.split('-').map(Number)
  return `${d}.${m}`
}

export function prettyDateLong(iso: string): string {
  const [y, m, d] = iso.split('-').map(Number)
  const date = new Date(y, m - 1, d)
  const weekday = WEEKDAYS_RU_FULL[date.getDay()]
  const month = MONTHS_RU[m - 1]
  return `${weekday} • ${d} ${month}`
}

// Returns { from: today, to: end of current week (Sunday) }
export function weekFromToday(base = new Date()): { from: string; to: string } {
  const from = new Date(base)
  const day = from.getDay()
  const daysUntilSunday = day === 0 ? 0 : 7 - day
  const to = new Date(from)
  to.setDate(to.getDate() + daysUntilSunday)
  return { from: isoDate(from), to: isoDate(to) }
}

export interface CalendarDay {
  iso: string
  dayName: string
  dayNumber: number
  monthName: string
  isToday: boolean
}

export function getWeekDays(base = new Date()): CalendarDay[] {
  const current = new Date(base)
  const dayOfWeek = current.getDay() === 0 ? 7 : current.getDay()
  const monday = new Date(current)
  monday.setDate(current.getDate() - dayOfWeek + 1)

  const todayIso = isoDate(new Date())

  const days: CalendarDay[] = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(monday)
    d.setDate(monday.getDate() + i)
    const iso = isoDate(d)
    days.push({
      iso,
      dayName: WEEKDAYS_RU[i],
      dayNumber: d.getDate(),
      monthName: MONTHS_RU_NOM[d.getMonth()],
      isToday: iso === todayIso,
    })
  }
  return days
}

