'use client'

import { useMemo, useState } from 'react'
import { useMe, useScheduleWeek, useClients } from '@/lib/hooks'
import { weekFromToday, WEEKDAYS_RU, prettyDateLong } from '@/lib/dates'
import { ScreenHeader, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { Card } from '@/components/ui/card'
import { SkeletonList } from '@/components/ui/skeleton'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import type { ScheduleEntry } from '@/lib/types'
import { Plus, Clock, Users, User, ChevronLeft, ChevronRight } from 'lucide-react'
import CreateTrainingModal from '@/components/create-training-modal'
import { Button } from '@/components/ui/button'

function LessonRow({ entry }: { entry: ScheduleEntry }) {
  const isCanceled = entry.status === 'canceled'

  return (
    <div className={`flex items-center gap-3 rounded-2xl p-3.5 transition-all duration-200 ${
      isCanceled
        ? 'bg-destructive/5 border border-destructive/20 opacity-75'
        : 'bg-card shadow-card border border-border/30'
    }`}>
      <Avatar className="h-10 w-10 border border-border/50 shadow-sm shrink-0">
        <AvatarFallback className={`text-xs font-bold uppercase ${
          isCanceled ? 'bg-destructive/10 text-destructive' : 'bg-primary/10 text-primary'
        }`}>
          {entry.group_name ? <Users className="h-4 w-4" /> : (entry.client_name?.charAt(0) || '?')}
        </AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-bold text-foreground truncate">
          {entry.group_name || entry.client_name}
        </p>
        <div className="flex items-center gap-1.5 mt-0.5">
          <Clock className="h-3 w-3 text-muted-foreground" />
          <span className="text-xs text-muted-foreground">{entry.time} · {entry.duration} мин</span>
        </div>
      </div>
      <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${
        isCanceled
          ? 'text-destructive bg-destructive/10'
          : 'text-success bg-success-light'
      }`}>
        {isCanceled ? 'Отменено' : 'Активно'}
      </span>
    </div>
  )
}

function CoachView() {
  const { from, to } = useMemo(() => weekFromToday(), [])
  const { data, isLoading, error } = useScheduleWeek(from, to)
  const [showCreate, setShowCreate] = useState(false)

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
    <>
      <CreateTrainingModal open={showCreate} onClose={() => setShowCreate(false)} />
      <div className="px-5 pb-24">
        {isLoading && <SkeletonList count={3} />}
        {error && <ErrorBox error={error} />}
        {!isLoading && byDay.length === 0 && <Empty text="На эту неделю тренировок нет" />}

        {byDay.map(([date, entries]) => (
          <div key={date} className="mb-6">
            <div className="mb-2.5 px-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">
              {prettyDateLong(date)}
            </div>
            <div className="flex flex-col gap-2">
              {entries.map((e) => (
                <LessonRow key={e.id} entry={e} />
              ))}
            </div>
          </div>
        ))}
      </div>

      <button
        onClick={() => setShowCreate(true)}
        className="fixed bottom-24 right-5 z-40 h-14 w-14 rounded-2xl bg-gradient-to-r from-primary to-cyan-500 text-white shadow-glow flex items-center justify-center active:scale-90 transition-transform"
      >
        <Plus className="h-6 w-6" />
      </button>
    </>
  )
}

function ParentView() {
  const { data: me } = useMe()
  const { from, to } = useMemo(() => weekFromToday(), [])
  const { data: schedule, isLoading } = useScheduleWeek(from, to)

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

  return (
    <div className="px-5 pb-24">
      {me?.children && me.children.length > 0 ? (
        <div className="mb-6">
          {me.children.map((c) => (
            <Card key={c.id} className="mb-3 flex items-center gap-3">
              <Avatar className="h-10 w-10 border border-border/50 shrink-0">
                <AvatarFallback className="bg-primary/10 text-primary text-xs font-bold uppercase">
                  {c.full_name.charAt(0)}
                </AvatarFallback>
              </Avatar>
              <div>
                <p className="text-sm font-bold text-foreground">{c.full_name}</p>
                <p className="text-xs text-muted-foreground">Расписание</p>
              </div>
            </Card>
          ))}
        </div>
      ) : (
        <Empty text="Привяжите ребёнка, чтобы видеть расписание" />
      )}

      {isLoading && <Spinner label="Загрузка..." />}
      {!isLoading && byDay.map(([date, entries]) => (
        <div key={date} className="mb-5">
          <div className="mb-2.5 px-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">
            {prettyDateLong(date)}
          </div>
          <div className="flex flex-col gap-2">
            {entries.map((e) => (
              <LessonRow key={e.id} entry={e} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

export default function Schedule() {
  const { data: me } = useMe()

  return (
    <div className="pt-2">
      {me?.role === 'parent' ? <ParentView /> : <CoachView />}
    </div>
  )
}
