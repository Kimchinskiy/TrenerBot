'use client'

import { useMemo, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useMe, useScheduleWeek, usePendingLeads } from '@/lib/hooks'
import { weekFromToday } from '@/lib/dates'
import { SkeletonList } from '@/components/ui/skeleton'
import { Card } from '@/components/ui/card'
import { Clock, ArrowRight, Bell, CheckCircle, Users, UserPlus } from 'lucide-react'
import Link from 'next/link'

function TodayLessons({ entries }: { entries: { time: string; client_name: string; group_name?: string | null; status: string; duration: number }[] }) {
  if (entries.length === 0) {
    return (
      <Card className="flex flex-col items-center justify-center p-6 text-center border-dashed border-border/60 bg-card/50">
        <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-muted mb-2 text-muted-foreground">
          <Clock className="h-6 w-6" />
        </div>
        <p className="text-sm font-semibold text-foreground">На сегодня тренировок нет</p>
        <p className="text-xs text-muted-foreground mt-0.5">Отличный день для отдыха или подготовки</p>
      </Card>
    )
  }

  if (entries.length > 5) {
    return (
      <Card className="p-5 text-center bg-card shadow-card border-border/40">
        <p className="text-base font-bold text-foreground mb-1">
          Запланировано {entries.length} тренировок на сегодня
        </p>
        <p className="text-xs text-muted-foreground mb-3">
          Посмотрите полный список и детали в расписании
        </p>
        <Link
          href="/dashboard/schedule"
          className="inline-flex items-center justify-center gap-2 rounded-xl bg-primary/10 px-4 py-2 text-sm font-bold text-primary active:scale-95 transition-transform"
        >
          Открыть расписание <ArrowRight className="h-4 w-4" />
        </Link>
      </Card>
    )
  }

  return (
    <div className="flex flex-col gap-2.5">
      {entries.map((e, i) => {
        const isCanceled = e.status === 'canceled'
        const isCompleted = e.status === 'completed'

        return (
          <div
            key={i}
            className={`flex items-center justify-between gap-3 rounded-2xl bg-card p-4 shadow-card border border-border/50 transition-all ${
              isCanceled ? 'opacity-60 bg-muted/30' : ''
            }`}
          >
            {/* Time Block */}
            <div className="flex items-center gap-3 min-w-0">
              <div className="flex flex-col items-center justify-center h-12 w-14 rounded-xl bg-primary/10 text-primary shrink-0 border border-primary/20">
                <span className="text-sm font-extrabold leading-none">{e.time}</span>
                <span className="text-[10px] font-medium text-primary/80 mt-1">{e.duration} мин</span>
              </div>

              {/* Details */}
              <div className="flex-1 min-w-0">
                <p className="text-base font-bold text-foreground truncate leading-snug">
                  {e.group_name || e.client_name}
                </p>
                {e.group_name && (
                  <p className="text-xs text-muted-foreground truncate mt-0.5">
                    Ученик: {e.client_name}
                  </p>
                )}
              </div>
            </div>

            {/* Status */}
            <div className="shrink-0">
              {isCanceled ? (
                <span className="text-[11px] font-bold text-destructive bg-destructive/10 px-2.5 py-1 rounded-full">
                  Отменено
                </span>
              ) : isCompleted ? (
                <span className="text-[11px] font-bold text-emerald-600 bg-emerald-50 dark:bg-emerald-950/40 px-2.5 py-1 rounded-full flex items-center gap-1">
                  <CheckCircle className="h-3 w-3" /> Проведено
                </span>
              ) : (
                <span className="text-[11px] font-bold text-primary bg-primary/10 px-2.5 py-1 rounded-full">
                  Запланировано
                </span>
              )}
            </div>
          </div>
        )
      })}
    </div>
  )
}

export default function HomeScreen() {
  const router = useRouter()
  const { data: me, isLoading: meLoading } = useMe()
  const { from, to } = useMemo(() => weekFromToday(), [])
  const { data: schedule, isLoading: schedLoading } = useScheduleWeek(from, to)
  const { data: pendingLeads } = usePendingLeads()
  const [greeting, setGreeting] = useState('Добрый день')

  useEffect(() => {
    const h = new Date().getHours()
    if (h < 12) setGreeting('Доброе утро')
    else if (h < 18) setGreeting('Добрый день')
    else setGreeting('Добрый вечер')
  }, [])

  const today = useMemo(() => {
    const d = new Date()
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
  }, [])

  const todayLessons = useMemo(() => {
    if (!schedule) return []
    return schedule.filter((e) => e.date === today).sort((a, b) => a.time.localeCompare(b.time))
  }, [schedule, today])

  const firstName = me?.client?.full_name?.split(' ')[0] || (me?.role === 'parent' ? 'Родитель' : 'Тренер')

  if (meLoading || schedLoading) {
    return (
      <div className="px-5 pt-6">
        <div className="h-8 w-48 shimmer rounded-xl mb-6" />
        <SkeletonList count={3} />
      </div>
    )
  }

  return (
    <div className="pb-24">
      {/* Header */}
      <div className="px-5 pt-6 pb-2">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-muted-foreground">{greeting}</p>
            <h1 className="text-display font-bold tracking-tight gradient-text mt-0.5">
              {firstName}!
            </h1>
          </div>
          <button
            onClick={() => router.push('/dashboard/notifications')}
            className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 active:scale-90 transition-transform"
          >
            <Bell className="h-5 w-5 text-primary" />
          </button>
        </div>
      </div>

      <div className="px-5 flex flex-col gap-5 pt-3">
        {/* Today's schedule */}
        <section>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-heading font-bold text-foreground">Сегодня</h2>
          </div>
          <TodayLessons entries={todayLessons} />
        </section>

        {/* Standalone Leads Section */}
        {(me?.role === 'admin' || me?.role === 'coach') && (
          <section>
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-heading font-bold text-foreground flex items-center gap-2">
                Заявки на запись
                {pendingLeads && pendingLeads.length > 0 && (
                  <span className="flex h-5 min-w-[1.25rem] items-center justify-center rounded-full bg-destructive text-[11px] font-bold text-destructive-foreground px-1.5">
                    {pendingLeads.length}
                  </span>
                )}
              </h2>
              <Link href="/dashboard/leads" className="text-xs font-semibold text-primary flex items-center gap-1">
                Все заявки <ArrowRight className="h-3 w-3" />
              </Link>
            </div>
            <Link href="/dashboard/leads">
              <Card className="flex items-center gap-4 py-4 px-5 bg-gradient-to-r from-primary/10 via-primary/5 to-transparent border-primary/20 hover:shadow-elevated transition-all">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary text-primary-foreground shrink-0 shadow-sm">
                  <UserPlus className="h-6 w-6" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-base font-bold text-foreground">Новые заявки</p>
                  <p className="text-xs text-muted-foreground">
                    {pendingLeads?.length
                      ? `${pendingLeads.length} ожидают рассмотрения`
                      : 'Новых заявок пока нет'}
                  </p>
                </div>
                <ArrowRight className="h-4 w-4 text-muted-foreground" />
              </Card>
            </Link>
          </section>
        )}

        {/* Quick actions */}
        <section>
          <h2 className="text-heading font-bold text-foreground mb-3">Быстрые действия</h2>
          <div className="flex flex-col gap-2">
            <Link href="/dashboard/attendance">
              <Card className="flex items-center gap-4 py-4 px-5 hover:shadow-elevated">
                <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 shrink-0">
                  <CheckCircle className="h-5 w-5 text-primary" />
                </div>
                <div className="flex-1">
                  <p className="text-base font-bold text-foreground">Посещаемость</p>
                  <p className="text-xs text-muted-foreground">Отметить посещения</p>
                </div>
                <ArrowRight className="h-4 w-4 text-muted-foreground" />
              </Card>
            </Link>
          </div>
        </section>
      </div>
    </div>
  )
}
