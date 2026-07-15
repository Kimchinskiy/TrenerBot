export function isoDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function weekRange(base = new Date()): { from: string; to: string } {
  const from = new Date(base)
  const day = from.getDay() === 0 ? 7 : from.getDay() // make Sunday = 7
  from.setDate(from.getDate() - day + 1) // Monday
  const to = new Date(from)
  to.setDate(to.getDate() + 7)
  return { from: isoDate(from), to: isoDate(to) }
}

export const WEEKDAYS_RU = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

export function prettyDate(iso: string): string {
  const [y, m, d] = iso.split('-').map(Number)
  return `${d}.${m}`
}
