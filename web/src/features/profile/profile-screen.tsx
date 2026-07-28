'use client'

import { useRouter } from 'next/navigation'
import { useMe, useCoachOnboarding, useUpgradeToCoach, useStartCoachTrial } from '@/lib/hooks'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox, Row } from '@/components/ui/screen'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import type { Client } from '@/lib/types'
import { Bell, Settings, HelpCircle, LogOut, Users, Calendar, BarChart3, ChevronRight, Crown, Palette } from 'lucide-react'
import { useAuth } from '@/components/auth-provider'
import { logout } from '@/lib/auth'
import { useState } from 'react'
import { ThemeSelector, ThemeToggle } from '@/components/theme-toggle'

function ClientCard({ c }: { c: Client }) {
  const isActive = c.status === 'active'
  const initials = c.full_name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase()

  return (
    <Card className="flex items-center gap-4">
      <Avatar className="h-14 w-14 border-2 border-white shadow-sm">
        <AvatarFallback className="bg-primary/10 text-primary text-lg font-bold uppercase">
          {initials}
        </AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0">
        <h2 className="text-lg font-bold text-foreground truncate">{c.full_name}</h2>
        <div className="flex items-center gap-2 mt-1">
          <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider ${
            isActive
              ? 'bg-success-light text-success'
              : 'bg-muted text-muted-foreground'
          }`}>
            {isActive ? 'Активен' : c.status}
          </span>
          {c.age && <span className="text-xs text-muted-foreground">{c.age} лет</span>}
        </div>
      </div>
    </Card>
  )
}

function CoachOnboarding() {
  const { data, isLoading } = useCoachOnboarding()
  const upgrade = useUpgradeToCoach()
  const [showForm, setShowForm] = useState(false)
  const [fullName, setFullName] = useState('')
  const [sport, setSport] = useState('')

  if (isLoading) return <Spinner label="Загрузка..." />

  if (!data?.is_coach) {
    if (!showForm) {
      return (
        <Card className="border-primary/20 relative overflow-hidden">
          <div className="absolute top-0 right-0 h-32 w-32 bg-primary/10 rounded-full blur-3xl" />
          <div className="flex flex-col items-center text-center gap-4 py-4 relative">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10">
              <Crown className="h-7 w-7 text-primary" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-foreground mb-1">Станьте тренером!</h3>
              <p className="text-sm text-muted-foreground">{data?.message || 'Ведёте учёт в заметках? Платформа Плавли создана от тренеров для тренеров.'}</p>
            </div>
            <Button onClick={() => setShowForm(true)} variant="gradient" className="w-full">
              Начать
            </Button>
            <p className="text-xs text-muted-foreground">7 дней бесплатно, затем 990₽/мес</p>
          </div>
        </Card>
      )
    }

    return (
      <Card>
        <h3 className="text-lg font-bold text-foreground mb-4">Регистрация тренера</h3>
        <div className="flex flex-col gap-3">
          <input
            className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-2.5 text-sm"
            placeholder="Ваше имя и фамилия"
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
          />
          <input
            className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-2.5 text-sm"
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
              variant="gradient"
              className="flex-1"
            >
              {upgrade.isPending ? 'Создание...' : 'Стать тренером'}
            </Button>
          </div>
          {upgrade.isSuccess && (
            <p className="text-center text-success text-sm font-semibold">Аккаунт тренера создан!</p>
          )}
          {upgrade.isError && (
            <p className="text-center text-destructive text-sm">{upgrade.error?.message}</p>
          )}
        </div>
      </Card>
    )
  }

  const sub = data.subscription
  const isActive = data.active
  const daysLeft = data.days_left ?? 0

  return (
    <Card className="border-primary/20 relative overflow-hidden">
      <div className="absolute top-0 right-0 h-32 w-32 bg-primary/10 rounded-full blur-3xl" />
      <div className="flex items-center gap-4 mb-3 relative">
        <Avatar className="h-14 w-14 border-2 border-primary shadow-sm">
          <AvatarFallback className="bg-primary/10 text-primary text-lg font-bold uppercase">
            {data.coach?.full_name?.charAt(0) || 'T'}
          </AvatarFallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <h2 className="text-lg font-bold text-foreground truncate">{data.coach?.full_name || 'Тренер'}</h2>
          <span className={`mt-1 inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider ${
            isActive
              ? 'bg-success-light text-success'
              : 'bg-destructive/10 text-destructive'
          }`}>
            {isActive ? (sub?.status === 'trial' ? 'Пробный' : 'Активна') : 'Неактивна'}
          </span>
        </div>
      </div>
      <div className="flex flex-col gap-2 relative">
        {sub?.status === 'trial' && (
          <Row label="Пробный период" value={`${daysLeft} дн.`} />
        )}
        {sub?.status === 'active' && sub.paid_until && (
          <Row label="Оплачено до" value={sub.paid_until} />
        )}
      </div>
    </Card>
  )
}

function MenuItem({ icon: Icon, label, onClick, variant }: { icon: React.ElementType; label: string; onClick: () => void; variant?: 'default' | 'danger' }) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-3.5 w-full rounded-2xl p-4 transition-all duration-200 active:scale-[0.99] ${
        variant === 'danger'
          ? 'bg-destructive/5 text-destructive hover:bg-destructive/10'
          : 'bg-card shadow-card border border-border/30 hover:shadow-elevated text-foreground'
      }`}
    >
      <div className={`flex h-9 w-9 items-center justify-center rounded-xl shrink-0 ${
        variant === 'danger' ? 'bg-destructive/10' : 'bg-primary/10'
      }`}>
        <Icon className={`h-4 w-4 ${variant === 'danger' ? 'text-destructive' : 'text-primary'}`} />
      </div>
      <span className="flex-1 text-left text-sm font-semibold">{label}</span>
      <ChevronRight className="h-4 w-4 text-muted-foreground/50 shrink-0" />
    </button>
  )
}

const UNDER_DEVELOPMENT = 'Раздел находится в разработке'

export default function Profile() {
  const router = useRouter()
  const { data, isLoading, error } = useMe()
  const { forceGuest } = useAuth()
  const [devMsg, setDevMsg] = useState('')

  const showDev = (label: string) => {
    setDevMsg(`«${label}» — ${UNDER_DEVELOPMENT.toLowerCase()}`)
    setTimeout(() => setDevMsg(''), 3000)
  }

  const handleLogout = async () => {
    await logout()
    forceGuest()
    router.replace('/login')
  }

  const displayName = data?.client?.full_name || (data?.role === 'admin' ? 'Администратор' : data?.role === 'coach' ? 'Тренер' : data?.role === 'parent' ? 'Родитель' : 'Клиент')
  const roleLabel = data?.role === 'coach' ? 'Тренер' : data?.role === 'admin' ? 'Администратор' : data?.role === 'parent' ? 'Родитель' : 'Клиент'

  return (
    <div className="pb-24 pt-6">
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}
      {!isLoading && !data && <Empty text="Профиль не найден" />}

      {data && (
        <div className="px-5 flex flex-col gap-5">
          {/* Profile header */}
          <div className="flex flex-col items-center">
            <Avatar className="h-20 w-20 border-4 border-card shadow-elevated mb-3">
              <AvatarFallback className="bg-primary/10 text-primary text-2xl font-bold">
                {displayName.charAt(0)}
              </AvatarFallback>
            </Avatar>
            <h2 className="text-title font-bold text-foreground">{displayName}</h2>
            <p className="text-sm text-muted-foreground mt-0.5">{roleLabel}</p>
          </div>

          {/* Coach onboarding */}
          {data.role === 'coach' && <CoachOnboarding />}

          {/* Dev message */}
          {devMsg && (
            <div className="rounded-2xl border border-border/30 bg-muted/50 px-4 py-3 text-center text-sm font-medium text-muted-foreground">
              {devMsg}
            </div>
          )}

          {/* Menu */}
          <div className="flex flex-col gap-2">
            {(data.role === 'coach' || data.role === 'admin') && (
              <>
                <MenuItem icon={Users} label="Мои клиенты" onClick={() => router.push('/dashboard/clients')} />
                <MenuItem icon={Calendar} label="Мой график" onClick={() => router.push('/dashboard/schedule')} />
                <MenuItem icon={BarChart3} label="Статистика" onClick={() => router.push('/dashboard/statistics')} />
                <MenuItem icon={Bell} label="Оповестить клиентов" onClick={() => router.push('/dashboard/notifications')} />
              </>
            )}
            <MenuItem icon={Settings} label="Настройки" onClick={() => showDev('Настройки')} />
            <MenuItem icon={HelpCircle} label="Поддержка" onClick={() => showDev('Поддержка')} />
          </div>

          {/* Theme Switcher Card */}
          <Card className="p-4 border-border/50 bg-card">
            <div className="flex items-center gap-3 mb-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary shrink-0">
                <Palette className="h-4.5 w-4.5" />
              </div>
              <div>
                <h3 className="text-sm font-bold text-foreground">Тема оформления</h3>
                <p className="text-xs text-muted-foreground">Выберите внешний вид приложения</p>
              </div>
            </div>
            <ThemeSelector />
          </Card>

          {/* Logout */}
          <MenuItem icon={LogOut} label="Выйти из аккаунта" onClick={handleLogout} variant="danger" />
        </div>
      )}
    </div>
  )
}
