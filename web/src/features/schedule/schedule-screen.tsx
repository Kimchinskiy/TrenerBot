'use client'

import { useMemo, useState } from 'react'
import { useMe, useScheduleWeek, useChildrenLessonStatus, useGroupClients } from '@/lib/hooks'
import { weekFromToday, prettyDateLong } from '@/lib/dates'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { LessonCard } from '@/components/lesson-card'
import type { ScheduleEntry, GroupMember } from '@/lib/types'
import { Play, Clock, CheckCircle2, XCircle, Users, X, Loader2 } from 'lucide-react'
import LinkChildSection from '@/features/parent/link-child-screen'
import CreateTrainingModal from '@/components/create-training-modal'
import { Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'

function StatusBadge({ status }: { status: { is_ongoing: boolean; minutes_left?: number | null; minutes_until?: number | null; has_lesson_today: boolean; time: string } }) {
  if (status.is_ongoing) {
    return (
      <div className="flex items-center gap-1.5 text-emerald-400 text-sm font-bold">
        <Play className="h-3.5 w-3.5 fill-emerald-400" />
        <span>Тренировка идёт</span>
        {status.minutes_left != null && <span className="text-muted-foreground font-normal">(осталось {status.minutes_left} мин)</span>}
      </div>
    )
  }
  if (status.minutes_until != null && status.minutes_until <= 60 && status.minutes_until > 0) {
    return (
      <div className="flex items-center gap-1.5 text-amber-400 text-sm font-bold">
        <Clock className="h-3.5 w-3.5" />
        <span>Скоро начнётся (через {status.minutes_until} мин)</span>
      </div>
    )
  }
  if (status.has_lesson_today) {
    return (
      <div className="flex items-center gap-1.5 text-blue-400 text-sm">
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

function ParentView() {
  const { data: me } = useMe()
  const { data: statuses } = useChildrenLessonStatus()
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

  return (
    <div className="px-4 pb-24">
      {me?.children && me.children.length > 0 ? (
        <div className="mb-6">
          <div className="mb-3 px-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">Статус детей</div>
          {me.children.map((c) => {
            const s = statuses?.find((st) => st.client_id === c.id)
            return (
              <Card key={c.id} className="mb-3 shadow-sm border-border/80">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold text-sm uppercase">
                    {c.full_name.charAt(0)}
                  </div>
                  <div className="flex-1">
                    <h3 className="font-bold text-foreground">{c.full_name}</h3>
                    {s ? <StatusBadge status={s} /> : <span className="text-muted-foreground text-sm">Загрузка...</span>}
                  </div>
                </div>
              </Card>
            )
          })}
        </div>
      ) : (
        <div className="mb-6">
          <p className="text-sm text-muted-foreground mb-3">Привяжите ребёнка, чтобы видеть его расписание</p>
        </div>
      )}

      <LinkChildSection />

      <div className="mt-6">
        <h3 className="text-sm font-bold tracking-wider text-muted-foreground uppercase mb-3 px-1">Расписание на неделю</h3>
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
  )
}

function CoachClientView({ role }: { role?: string }) {
  const { from, to } = useMemo(() => weekFromToday(), [])
  const { data, isLoading, error } = useScheduleWeek(from, to)
  const [showCreate, setShowCreate] = useState(false)
  const [selectedGroup, setSelectedGroup] = useState<{ id: number; name: string } | null>(null)
  const isCoach = role === 'coach' || role === 'admin'
  const { data: groupMembers, isLoading: membersLoading } = useGroupClients(selectedGroup?.id || 0)

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
      <div className="px-4 pb-24">
        {isLoading && <Spinner label="Загрузка..." />}
        {error && <ErrorBox error={error} />}
        {!isLoading && byDay.length === 0 && <Empty text="На эту неделю тренировок нет" />}
        {byDay.map(([date, entries]) => (
          <div key={date} className="mb-6">
            <div className="mb-3 px-1 text-xs font-bold uppercase tracking-wider text-muted-foreground">{prettyDateLong(date)}</div>
            <div className="flex flex-col gap-1">
              {entries.map((e) => (
                <LessonCard
                  key={e.id}
                  entry={e}
                  onClick={e.group_id ? () => setSelectedGroup({ id: e.group_id!, name: e.group_name || 'Группа' }) : undefined}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
      {isCoach && (
        <button
          onClick={() => setShowCreate(true)}
          className="fixed bottom-24 right-5 z-40 h-14 w-14 rounded-full bg-primary text-primary-foreground shadow-xl flex items-center justify-center active:scale-90 transition-transform"
        >
          <Plus className="h-7 w-7" />
        </button>
      )}

      {selectedGroup && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" onClick={() => setSelectedGroup(null)}>
          <div className="w-full max-w-lg rounded-2xl bg-background p-5 shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-bold flex items-center gap-2"><Users className="h-5 w-5" /> {selectedGroup.name}</h2>
              <button onClick={() => setSelectedGroup(null)} className="rounded-full p-2 hover:bg-muted/60"><X className="h-5 w-5" /></button>
            </div>
            {membersLoading ? (
              <div className="flex items-center justify-center py-6"><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></div>
            ) : groupMembers && groupMembers.length > 0 ? (
              <div className="flex flex-col gap-2 max-h-64 overflow-y-auto">
                {groupMembers.map((m) => (
                  <div key={m.client_id} className="flex items-center justify-between rounded-xl border border-border bg-background px-4 py-3">
                    <span className="font-medium text-sm">{m.client_name}</span>
                    <span className="text-xs text-muted-foreground capitalize">{m.role}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">В этой группе пока нет участников</p>
            )}
            <div className="mt-4">
              <Button className="w-full" onClick={() => { setSelectedGroup(null); window.location.href = `/dashboard/groups/${selectedGroup.id}` }}>Открыть группу</Button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default function Schedule() {
  const { data: me } = useMe()

  return (
    <div>
      <ScreenHeader title={me?.role === 'parent' ? 'Дети' : 'Расписание'} subtitle={me?.role === 'parent' ? '' : 'Эта неделя'} />
      {me?.role === 'parent' ? <ParentView /> : <CoachClientView role={me?.role} />}
    </div>
  )
}
