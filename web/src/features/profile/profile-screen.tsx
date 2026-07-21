'use client'

import { useRouter } from 'next/navigation'
import { useMe, useCoachOnboarding, useUpgradeToCoach, useStartCoachTrial, useUpgradeToParent } from '@/lib/hooks'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox, Row } from '@/components/ui/screen'
import { Button } from '@/components/ui'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import type { Client } from '@/lib/types'
import { Bell, Sparkles, Timer, Crown, LogOut } from 'lucide-react'
import { useState } from 'react'
import { useAuth } from '@/components/auth-provider'
import { logout } from '@/lib/auth'

function ClientCard({ c }: { c: Client }) {
  const isActive = c.status === 'active'

  return (
    <Card className="mb-4 shadow-md border-border/80 relative overflow-hidden">
      <div className="absolute top-0 right-0 h-24 w-24 bg-primary/10 rounded-full blur-3xl" />
      <div className="flex items-center gap-4 mb-4">
        <Avatar className="h-14 w-14 border-2 border-border shadow-sm">
          <AvatarFallback className="bg-primary/10 text-primary text-xl font-bold uppercase">
            {c.full_name.charAt(0)}
          </AvatarFallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <h2 className="text-xl font-bold text-foreground truncate">{c.full_name}</h2>
          <div className="mt-1 flex items-center">
            <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-bold uppercase tracking-wider border ${
              isActive
                ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400'
                : 'border-muted-foreground/30 bg-muted text-muted-foreground'
            }`}>
              {isActive ? 'Активен' : c.status}
            </span>
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-1.5 pt-2">
        <Row label="Телефон" value={c.phone} />
        <Row label="Возраст" value={c.age ? `${c.age} лет` : null} />
        <Row label="Мед. ограничения" value={c.medical_limits} />
        {c.subscription_ends_at && <Row label="Подписка до" value={c.subscription_ends_at} />}
      </div>
    </Card>
  )
}

function CoachOnboarding() {
  const router = useRouter()
  const { data, isLoading } = useCoachOnboarding()
  const upgrade = useUpgradeToCoach()
  const startTrial = useStartCoachTrial()
  const [showForm, setShowForm] = useState(false)
  const [fullName, setFullName] = useState('')
  const [sport, setSport] = useState('')

  if (isLoading) return <Spinner label="Загрузка..." />

  if (!data?.is_coach) {
    if (!showForm) {
      return (
        <Card className="mb-4 shadow-md border-primary/30 relative overflow-hidden">
          <div className="absolute top-0 right-0 h-32 w-32 bg-primary/20 rounded-full blur-3xl" />
          <div className="flex flex-col items-center text-center gap-4 py-4">
            <Crown className="h-10 w-10 text-primary" />
            <div>
              <h3 className="text-lg font-bold mb-2">Станьте тренером!</h3>
              <p className="text-sm text-muted-foreground">{data?.message || 'До сих пор ведете учет в заметках или Excel? Платформа Плавли создана от тренеров для тренеров.'}</p>
            </div>
            <Button onClick={() => setShowForm(true)} className="w-full font-bold shadow-md">
              Начать
            </Button>
            <p className="text-xs text-muted-foreground">7 дней бесплатно, затем 990₽/мес</p>
          </div>
        </Card>
      )
    }

    return (
      <Card className="mb-4 shadow-md">
        <h3 className="text-lg font-bold mb-4">Регистрация тренера</h3>
        <div className="flex flex-col gap-3">
          <input
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
            placeholder="Ваше имя и фамилия"
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
          />
          <input
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
            placeholder="Вид спорта (необязательно)"
            value={sport}
            onChange={(e) => setSport(e.target.value)}
          />
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setShowForm(false)} className="flex-1">Назад</Button>
            <Button
              onClick={() => {
                if (!fullName.trim()) return
                upgrade.mutate({ fullName: fullName.trim(), sport: sport.trim() })
              }}
              disabled={upgrade.isPending || !fullName.trim()}
              className="flex-1 font-bold shadow-md"
            >
              {upgrade.isPending ? 'Создание...' : 'Стать тренером'}
            </Button>
          </div>
          {upgrade.isSuccess && (
            <div className="text-center text-emerald-400 text-sm font-bold">
              Аккаунт тренера создан! 7 дней бесплатно.
            </div>
          )}
          {upgrade.isError && (
            <div className="text-center text-red-400 text-sm">Ошибка: {upgrade.error?.message}</div>
          )}
        </div>
      </Card>
    )
  }

  // Coach with subscription info
  const sub = data.subscription
  const isActive = data.active
  const daysLeft = data.days_left ?? 0

  return (
    <Card className="mb-4 shadow-md border-primary/30 relative overflow-hidden">
      <div className="absolute top-0 right-0 h-32 w-32 bg-primary/10 rounded-full blur-3xl" />
      <div className="flex items-center gap-4 mb-4">
        <Avatar className="h-14 w-14 border-2 border-primary shadow-sm">
          <AvatarFallback className="bg-primary/10 text-primary text-xl font-bold uppercase">
            {data.coach?.full_name?.charAt(0) || 'T'}
          </AvatarFallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <h2 className="text-xl font-bold text-foreground truncate">{data.coach?.full_name || 'Тренер'}</h2>
          <div className="mt-1 flex items-center gap-2">
            <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-bold uppercase tracking-wider border ${
              isActive
                ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400'
                : 'border-red-500/30 bg-red-500/10 text-red-400'
            }`}>
              {isActive ? (sub?.status === 'trial' ? 'Пробный' : 'Активна') : 'Неактивна'}
            </span>
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-2 pt-2">
        {sub?.status === 'trial' && (
          <>
            <Row label="Пробный период" value={`${daysLeft} дн.`} />
            {daysLeft <= 0 ? (
              <Button onClick={() => startTrial.mutate()} disabled={startTrial.isPending} className="w-full mt-2 font-bold shadow-md">
                {startTrial.isPending ? 'Загрузка...' : 'Активировать подписку'}
              </Button>
            ) : (
              <div className="flex items-center gap-2 text-sm text-muted-foreground mt-1">
                <Timer className="h-4 w-4" />
                <span>Осталось {daysLeft} дн. бесплатно</span>
              </div>
            )}
          </>
        )}
        {sub?.status === 'active' && sub.paid_until && (
          <Row label="Оплачено до" value={sub.paid_until} />
        )}
        {!isActive && (
          <div className="mt-2 space-y-2">
            <p className="text-sm text-muted-foreground">Подписка неактивна. Данные доступны только для просмотра.</p>
            <Button onClick={() => router.push('/dashboard/more/subscriptions')} className="w-full font-bold shadow-md">
              Оформить подписку
            </Button>
          </div>
        )}
      </div>
    </Card>
  )
}

function ParentSection({ role }: { role: string }) {
  const router = useRouter()
  const upgrade = useUpgradeToParent()

  if (role === 'parent') {
    return (
      <Button
        onClick={() => router.push('/dashboard/schedule')}
        className="w-full flex items-center justify-center gap-2 font-bold shadow-md"
      >
        <Sparkles className="h-5 w-5" />
        <span>Расписание детей</span>
      </Button>
    )
  }

  return (
    <Card className="mb-4 shadow-md border-border/80">
      <div className="flex flex-col items-center text-center gap-3 py-3">
        <Sparkles className="h-8 w-8 text-primary" />
        <div>
          <h3 className="text-base font-bold mb-1">Я родитель</h3>
          <p className="text-sm text-muted-foreground">Следите за тренировками ребёнка</p>
        </div>
        <Button
          onClick={() => upgrade.mutate()}
          disabled={upgrade.isPending}
          className="w-full font-bold shadow-md"
        >
          {upgrade.isPending ? 'Загрузка...' : 'Стать родителем'}
        </Button>
        {upgrade.isSuccess && (
          <p className="text-emerald-400 text-sm">Теперь привяжите ребёнка</p>
        )}
      </div>
    </Card>
  )
}

export default function Profile() {
  const router = useRouter()
  const { data, isLoading, error } = useMe()
  const { forceGuest } = useAuth()

  const handleLogout = async () => {
    await logout()
    forceGuest()
    router.replace('/login')
  }

  return (
    <div>
      <ScreenHeader title="Профиль" />
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}
      {!isLoading && !data && <Empty text="Профиль не найден" />}
      {data && (
        <div className="px-4 pb-24">
          {data.client && <ClientCard c={data.client} />}

          {data.role === 'coach' && <CoachOnboarding />}

          {data.role !== 'admin' && <ParentSection role={data.role} />}

          {data.role === 'parent' && data.children && data.children.length > 0 && (
            <div className="mt-4">
              <div className="mb-3 px-1 text-sm font-bold tracking-wider text-muted-foreground uppercase flex items-center gap-1.5">
                <Sparkles className="h-4 w-4 text-primary" />
                <span>Дети</span>
              </div>
              {data.children.map((c) => <ClientCard key={c.id} c={c} />)}
            </div>
          )}
          {(data.role === 'coach' || data.role === 'admin') && (
            <div className="mt-6">
              <Button
                onClick={() => router.push('/dashboard/more/notifications')}
                className="w-full flex items-center justify-center gap-2 font-bold shadow-md"
              >
                <Bell className="h-5 w-5" />
                <span>Оповестить клиентов</span>
              </Button>
            </div>
          )}

          <div className="mt-8 pt-6 border-t border-border/50">
            <Button
              onClick={handleLogout}
              variant="outline"
              className="w-full flex items-center justify-center gap-2 font-bold text-red-400 border-red-400/30 hover:bg-red-500/10"
            >
              <LogOut className="h-5 w-5" />
              <span>Выйти из аккаунта</span>
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
