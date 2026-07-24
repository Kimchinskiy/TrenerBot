'use client'

import { useMemo } from 'react'
import { useRouter } from 'next/navigation'
import { useMe, useScheduleWeek } from '@/lib/hooks'
import { weekFromToday } from '@/lib/dates'
import { WaveDivider } from '@/components/ui/wave-divider'
import { SkeletonList } from '@/components/ui/skeleton'
import { Card } from '@/components/ui/card'
import { Calendar, Clock, ArrowRight, Bell, CheckCircle, Users } from 'lucide-react'
import Link from 'next/link'

function getGreeting(): string {
  const h = new Date().getHours()
  if (h < 12) return 'Доброе утро'
  if (h < 18) return 'Добрый день'
  return 'Добрый вечер'
}

function TodayLessons({ entries }: { entries: { time: string; client_name: string; group_name?: string | null; status: string; duration: number }[] }) {
  if (entries.length === 0) {
    return (
      <div className="text-center py-6">
        <p className="text-sm text-muted-foreground">На сегодня тренировок нет</p>
      </div>
    )
  }

  if (entries.length > 5) {
    return (
      <div className="rounded-2xl bg-white p-5 shadow-card border border-border/30 text-center">
        <p className="text-base font-bold text-foreground mb-1">
          Сегодня запланировано {entries.length} тренировок
        </p>
        <Link
          href="/dashboard/schedule"
          className="inline-flex items-center gap-1 mt-2 text-sm font-semibold text-primary"
        >
          Перейти к расписанию <ArrowRight className="h-3 w-3" />
        </Link>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      {entries.map((e, i) => (
        <div
          key={i}
          className="flex items-center gap-3 rounded-2xl bg-white p-3.5 shadow-card border border-border/30"
        >
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary shrink-0">
            <Clock className="h-4 w-4" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="text-sm font-bold text-foreground">{e.time}</span>
              <span className="text-xs text-muted-foreground">{e.duration} мин</span>
            </div>
            <p className="text-sm font-medium text-muted-foreground truncate">
              {e.group_name || e.client_name}
            </p>
          </div>
          {e.status === 'canceled' && (
            <span className="text-[10px] font-bold uppercase tracking-wider text-destructive bg-destructive/10 px-2 py-0.5 rounded-full">
              Отменено
            </span>
          )}
        </div>
      ))}
    </div>
  )
}

export default function HomeScreen() {
  const router = useRouter()
  const { data: me, isLoading: meLoading } = useMe()
  const { from, to } = useMemo(() => weekFromToday(), [])
  const { data: schedule, isLoading: schedLoading } = useScheduleWeek(from, to)

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

  const firstName = me?.client?.full_name?.split(' ')[0] || me?.role === 'parent' ? 'Родитель' : 'Тренер'

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
            <p className="text-sm font-medium text-muted-foreground">{getGreeting()}</p>
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

      <WaveDivider className="text-primary/5 -mb-1" />

      <div className="px-5 flex flex-col gap-5">
        {/* Today's schedule */}
        <section>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-heading font-bold text-foreground">Сегодня</h2>
            <Link href="/dashboard/schedule" className="text-xs font-semibold text-primary flex items-center gap-1">
              Расписание <ArrowRight className="h-3 w-3" />
            </Link>
          </div>
          <TodayLessons entries={todayLessons} />
        </section>

        {/* Quick actions */}
        <section>
          <h2 className="text-heading font-bold text-foreground mb-3">Быстрые действия</h2>
          <div className="flex flex-col gap-2">
            <Link href="/dashboard/schedule">
              <Card className="flex items-center gap-4 py-4 px-5 hover:shadow-elevated">
                <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 shrink-0">
                  <Calendar className="h-5 w-5 text-primary" />
                </div>
                <div className="flex-1">
                  <p className="text-base font-bold text-foreground">Расписание</p>
                  <p className="text-xs text-muted-foreground">Просмотреть тренировки</p>
                </div>
                <ArrowRight className="h-4 w-4 text-muted-foreground" />
              </Card>
            </Link>
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
