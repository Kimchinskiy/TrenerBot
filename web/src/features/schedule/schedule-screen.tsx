'use client'

import { useMemo } from 'react'
import { useScheduleWeek } from '@/lib/hooks'
import { weekFromToday, prettyDateLong } from '@/lib/dates'
import { ScreenHeader, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { LessonCard } from '@/components/lesson-card'
import type { ScheduleEntry } from '@/lib/types'

export default function Schedule() {
  const { from, to } = useMemo(() => weekFromToday(), [])
  const { data, isLoading, error } = useScheduleWeek(from, to)

  const byDay = useMemo(() => {
    const map = new Map<string, ScheduleEntry[]>()
    ;(data || []).forEach((e) => {
      const arr = map.get(e.date) || []
      arr.push(e)
      map.set(e.date, arr)
    })
    return [...map.entries()].sort((a, b) => a[0].localeCompare(b[0]))
  }, [data])

  return (
    <div>
      <ScreenHeader title="Расписание" subtitle="Эта неделя" />
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}
      {!isLoading && byDay.length === 0 && <Empty text="На эту неделю тренировок нет" />}
      <div className="flex flex-col gap-3 px-4 pb-24">
        {byDay.map(([date, entries]) => (
          <div key={date} className="mb-6">
            <div className="mb-3 px-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">{prettyDateLong(date)}</div>
            <div className="flex flex-col gap-1">
              {entries.map((e) => (
                <LessonCard key={e.id} entry={e} />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
