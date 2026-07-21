'use client'

import { useMe, useChildrenLessonStatus } from '@/lib/hooks'
import { useScheduleWeek } from '@/lib/hooks'
import { weekFromToday, prettyDateLong } from '@/lib/dates'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { LessonCard } from '@/components/lesson-card'
import type { ScheduleEntry, ChildLessonStatus } from '@/lib/types'
import { useMemo } from 'react'
import { Play, Clock, CheckCircle2, XCircle } from 'lucide-react'
import LinkChildSection from './link-child-screen'

function StatusBadge({ status }: { status: ChildLessonStatus }) {
  if (status.is_ongoing) {
    return (
      <div className="flex items-center gap-1.5 text-success text-sm font-bold">
        <Play className="h-3.5 w-3.5 fill-success" />
        <span>Тренировка идёт</span>
        {status.minutes_left != null && <span className="text-muted-foreground font-normal">(осталось {status.minutes_left} мин)</span>}
      </div>
    )
  }
  if (status.minutes_until != null && status.minutes_until <= 60 && status.minutes_until > 0) {
    return (
      <div className="flex items-center gap-1.5 text-warning text-sm font-bold">
        <Clock className="h-3.5 w-3.5" />
        <span>Скоро начнётся (через {status.minutes_until} мин)</span>
      </div>
    )
  }
  if (status.has_lesson_today) {
    return (
      <div className="flex items-center gap-1.5 text-primary text-sm">
        <CheckCircle2 className="h-3.5 w-3.5" />
        <span>Тренировка сегодня в {status.time}</span>
      </div>
    )
  }
  return (
    <div className="flex items-center gap-1.5 text-muted-foreground text-sm">
      <XCircle className="h-3.5 w-3.5" />
      <span>Сегодня тренировок нет</span>
    </div>
  )
}

function ChildCard({ childId, fullName }: { childId: number; fullName: string }) {
  const { data: statuses } = useChildrenLessonStatus()
  const myStatus = statuses?.find((s) => s.client_id === childId)

  return (
    <Card className="mb-3 shadow-sm border-border/80 relative overflow-hidden">
      <div className="absolute top-0 right-0 h-20 w-20 bg-primary/5 rounded-full blur-2xl" />
      <div className="flex items-center gap-3 mb-3">
        <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold text-sm uppercase">
          {fullName.charAt(0)}
        </div>
        <div className="flex-1">
          <h3 className="font-bold text-foreground">{fullName}</h3>
          {myStatus && <StatusBadge status={myStatus} />}
          {!myStatus && <span className="text-muted-foreground text-sm">Загрузка...</span>}
        </div>
      </div>
      {myStatus?.has_lesson_today && !myStatus.is_ongoing && myStatus.minutes_until == null && (
        <div className="text-xs text-muted-foreground">
          {myStatus.time}, {myStatus.duration} мин
        </div>
      )}
    </Card>
  )
}

export default function ParentDashboard() {
  const { data: me, isLoading, error } = useMe()
  const { from, to } = useMemo(() => weekFromToday(), [])
  const { data: schedule, isLoading: schedLoading } = useScheduleWeek(from, to)

  const byDay = useMemo(() => {
    if (!schedule) return []
    const map = new Map<string, ScheduleEntry[]>()
    schedule.forEach((e) => {
      const arr = map.get(e.date) || []
      arr.push(e)
      map.set(e.date, arr)
    })
    return [...map.entries()].sort((a, b) => a[0].localeCompare(b[0]))
  }, [schedule])

  if (isLoading) return <Spinner label="Загрузка..." />
  if (error) return <ErrorBox error={error} />
  if (!me || me.role !== 'parent') return null

  return (
    <div>
      <ScreenHeader title="Дети" subtitle="Статус тренировок" />

      <div className="px-4 pb-24">
        {/* Children status cards */}
        {me.children && me.children.length > 0 ? (
          <div className="mb-6">
            {me.children.map((c) => (
              <ChildCard key={c.id} childId={c.id} fullName={c.full_name} />
            ))}
          </div>
        ) : (
          <div className="mb-6">
            <p className="text-sm text-muted-foreground mb-3">Привяжите ребёнка, чтобы видеть его расписание</p>
          </div>
        )}

        {/* Link child section */}
        <LinkChildSection />

        {/* Weekly schedule for all children */}
        <div className="mt-6">
          <h3 className="text-sm font-bold tracking-wider text-muted-foreground uppercase mb-3">Расписание на неделю</h3>
          {schedLoading && <Spinner label="Загрузка..." />}
          {!schedLoading && byDay.length === 0 && <Empty text="На эту неделю тренировок нет" />}
          {byDay.map(([date, entries]) => (
            <div key={date} className="mb-4">
              <div className="mb-2 px-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">{prettyDateLong(date)}</div>
              <div className="flex flex-col gap-1">
                {entries.map((e) => (
                  <LessonCard key={e.id} entry={e} />
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
