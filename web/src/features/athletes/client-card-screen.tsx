'use client'

import { useRouter } from 'next/navigation'
import { useClients } from '@/lib/hooks'
import { ScreenHeader, Card, Spinner, Empty, Row } from '@/components/ui/screen'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { WaveDivider } from '@/components/ui/wave-divider'
import { Phone, MessageCircle, Calendar, FileText, ArrowRight } from 'lucide-react'

export default function ClientCardScreen({ params }: { params: { id: string } }) {
  const router = useRouter()
  const clientId = parseInt(params.id, 10)
  const { data: clients, isLoading } = useClients()

  const client = clients?.find((c) => c.id === clientId)

  if (isLoading) {
    return (
      <div>
        <ScreenHeader title="Клиент" onBack={() => router.back()} />
        <Spinner label="Загрузка..." />
      </div>
    )
  }

  if (!client) {
    return (
      <div>
        <ScreenHeader title="Клиент" onBack={() => router.back()} />
        <Empty text="Клиент не найден" />
      </div>
    )
  }

  const isActive = client.status === 'active'
  const initials = client.full_name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase()

  return (
    <div className="pb-24">
      <ScreenHeader title="Клиент" onBack={() => router.back()} />

      {/* Avatar + Name */}
      <div className="px-5 flex flex-col items-center mb-2">
        <Avatar className="h-20 w-20 border-4 border-white shadow-elevated mb-3">
          <AvatarFallback className="bg-primary/10 text-primary text-2xl font-bold">
            {initials}
          </AvatarFallback>
        </Avatar>
        <h2 className="text-title font-bold text-foreground">{client.full_name}</h2>
        {client.age && (
          <p className="text-sm text-muted-foreground mt-0.5">{client.age} лет</p>
        )}
        <span className={`mt-2 inline-flex items-center rounded-full px-3 py-1 text-xs font-bold uppercase tracking-wider ${
          isActive
            ? 'bg-success-light text-success'
            : 'bg-muted text-muted-foreground'
        }`}>
          {isActive ? 'Активен' : client.status}
        </span>
      </div>

      <WaveDivider className="text-primary/5 -mb-1 mt-2" />

      <div className="px-5 flex flex-col gap-4">
        {/* Contact */}
        <section>
          <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">Контакты</h3>
          <Card className="py-0 divide-y divide-border/40">
            {client.phone && (
              <a href={`tel:${client.phone}`} className="flex items-center gap-3 p-4 hover:bg-muted/30 transition-colors">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 shrink-0">
                  <Phone className="h-4 w-4 text-primary" />
                </div>
                <div className="flex-1">
                  <p className="text-xs text-muted-foreground">Телефон</p>
                  <p className="text-sm font-semibold text-foreground">{client.phone}</p>
                </div>
                <ArrowRight className="h-4 w-4 text-muted-foreground/50" />
              </a>
            )}
            <div className="flex items-center gap-3 p-4">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 shrink-0">
                <MessageCircle className="h-4 w-4 text-primary" />
              </div>
              <div className="flex-1">
                <p className="text-xs text-muted-foreground">Telegram</p>
                <p className="text-sm font-semibold text-foreground">
                  {client.bot_access ? 'Подключён' : 'Не подключён'}
                </p>
              </div>
            </div>
          </Card>
        </section>

        {/* Subscription */}
        <section>
          <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">Абонемент</h3>
          <Card>
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 shrink-0">
                <Calendar className="h-4 w-4 text-primary" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Действует до</p>
                <p className="text-sm font-semibold text-foreground">
                  {client.subscription_ends_at || 'Не оформлен'}
                </p>
              </div>
            </div>
          </Card>
        </section>

        {/* Medical limits */}
        {client.medical_limits && (
          <section>
            <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">Медицинские ограничения</h3>
            <Card>
              <div className="flex items-start gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-warning-light shrink-0">
                  <FileText className="h-4 w-4 text-warning" />
                </div>
                <p className="text-sm text-foreground">{client.medical_limits}</p>
              </div>
            </Card>
          </section>
        )}

        {/* Schedule */}
        <section>
          <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">Расписание</h3>
          <Card className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">Просмотреть расписание клиента</p>
            <ArrowRight className="h-4 w-4 text-muted-foreground/50" />
          </Card>
        </section>

        {/* Message button */}
        <Button variant="gradient" size="lg" className="mt-2">
          <MessageCircle className="h-5 w-5 mr-2" />
          Написать клиенту
        </Button>
      </div>
    </div>
  )
}
