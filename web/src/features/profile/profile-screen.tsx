'use client'

import { useRouter } from 'next/navigation'
import { useMe } from '@/lib/hooks'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox, Row } from '@/components/ui/screen'
import { Button } from '@/components/ui'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import type { Client } from '@/lib/types'
import { Bell, Sparkles } from 'lucide-react'

function ClientCard({ c }: { c: Client }) {
  const isActive = c.status === 'active'

  return (
    <Card className="mb-4 shadow-md border-border/80 relative overflow-hidden">
      {/* Decorative gradient overlay */}
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
            <span
              className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-bold uppercase tracking-wider border ${
                isActive
                  ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400'
                  : 'border-muted-foreground/30 bg-muted text-muted-foreground'
              }`}
            >
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

export default function Profile() {
  const router = useRouter()
  const { data, isLoading, error } = useMe()

  return (
    <div>
      <ScreenHeader title="Профиль" />
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}
      {!isLoading && !data && <Empty text="Профиль не найден" />}
      {data && (
        <div className="px-4 pb-24">
          {data.client && <ClientCard c={data.client} />}
          {data.role === 'parent' && (
            <div className="mt-4">
              <div className="mb-3 px-1 text-sm font-bold tracking-wider text-muted-foreground uppercase flex items-center gap-1.5">
                <Sparkles className="h-4 w-4 text-primary" />
                <span>Дети</span>
              </div>
              {data.children && data.children.length > 0 ? (
                data.children.map((c) => <ClientCard key={c.id} c={c} />)
              ) : (
                <Empty text="Нет привязанных детей" />
              )}
            </div>
          )}
          {(data.role === 'coach' || data.role === 'admin') && (
            <div className="mt-6">
              <Button
                onClick={() => router.push('/dashboard/more/notifications')}
                className="flex items-center justify-center gap-2 font-bold shadow-md"
              >
                <Bell className="h-5 w-5" />
                <span>📢 Оповестить клиентов</span>
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
