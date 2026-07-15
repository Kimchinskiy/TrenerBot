import { useMemo } from 'react'
import { useMe, useSchedule } from '../lib/hooks'
import { weekRange, prettyDate } from '../lib/dates'
import { Card, ScreenHeader, Spinner, Empty, ErrorBox } from '../components/ui'
import type { Lesson } from '../lib/types'

function statusLabel(s: string) {
  switch (s) {
    case 'canceled':
      return 'Отменена'
    case 'moved':
      return 'Перенесена'
    case 'done':
      return 'Завершена'
    default:
      return 'Запланирована'
  }
}

export default function Schedule() {
  const me = useMe()
  const { from, to } = useMemo(() => weekRange(), [])
  const { data, isLoading, error } = useSchedule(from, to)

  const byDay = useMemo(() => {
    const map = new Map<string, Lesson[]>()
    ;(data || []).forEach((l) => {
      const arr = map.get(l.date) || []
      arr.push(l)
      map.set(l.date, arr)
    })
    return [...map.entries()].sort((a, b) => a[0].localeCompare(b[0]))
  }, [data])

  const isCoach = me.data?.role === 'coach'

  return (
    <div>
      <ScreenHeader
        title="Расписание"
        subtitle={isCoach ? 'Ваши тренировки' : 'Ближайшие 7 дней'}
      />
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}
      {!isLoading && byDay.length === 0 && <Empty text="На ближайшую неделю тренировок нет 🗓" />}
      <div className="flex flex-col gap-3 px-4 pb-24">
        {byDay.map(([date, lessons]) => (
          <div key={date}>
            <div className="mb-1 px-1 text-sm font-semibold text-tg-hint">{prettyDate(date)}</div>
            {lessons.map((l) => (
              <Card key={l.id} className="mb-2 flex items-center justify-between">
                <div>
                  <div className="text-lg font-semibold">{l.time}</div>
                  {l.location && <div className="text-sm text-tg-hint">📍 {l.location}</div>}
                </div>
                <div className="text-xs text-tg-hint">{statusLabel(l.status)}</div>
              </Card>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}
