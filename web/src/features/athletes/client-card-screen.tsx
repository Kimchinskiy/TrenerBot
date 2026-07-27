'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useClients, useClientSubscriptions, useCreateClientSubscription, useUpdateClientSubscription, useDeleteClientSubscription } from '@/lib/hooks'
import { ScreenHeader, Card, Spinner, Empty } from '@/components/ui/screen'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { WaveDivider } from '@/components/ui/wave-divider'
import { Phone, MessageCircle, Calendar, FileText, ArrowRight, Plus, Trash2, Check, CreditCard } from 'lucide-react'
import type { ClientSubscription } from '@/lib/types'

function SubscriptionCard({ sub, onDelete }: { sub: ClientSubscription; onDelete: (id: number) => void }) {
  const typeLabel = sub.type === 'count' ? `${sub.lessons_left} занятий` : sub.type === 'period' ? 'Безлимит' : sub.type
  const now = new Date()
  const end = new Date(sub.ends_at)
  const isExpired = end < now

  return (
    <Card className="flex items-center gap-3">
      <div className="flex h-9 w-9 items-center justify-center rounded-xl shrink-0 bg-primary/10">
        <CreditCard className="h-4 w-4 text-primary" />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-bold text-foreground">{typeLabel}</span>
          <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${
            isExpired ? 'bg-destructive/10 text-destructive' : 'bg-success-light text-success'
          }`}>
            {isExpired ? 'Истёк' : 'Активен'}
          </span>
        </div>
        <div className="text-xs text-muted-foreground mt-0.5">
          {sub.price > 0 && <span>{sub.price}₽ · </span>}
          до {sub.ends_at}
          {sub.lessons_left > 0 && <span> · {sub.lessons_left} ост.</span>}
        </div>
      </div>
      <button
        onClick={() => onDelete(sub.id)}
        className="rounded-xl p-2 hover:bg-destructive/10 transition-colors shrink-0"
      >
        <Trash2 className="h-4 w-4 text-destructive" />
      </button>
    </Card>
  )
}

function CreateSubscriptionForm({
  clientId,
  onDone,
}: {
  clientId: number
  onDone: () => void
}) {
  const create = useCreateClientSubscription()
  const [subType, setSubType] = useState<'count' | 'period'>('count')
  const [lessons, setLessons] = useState('8')
  const [price, setPrice] = useState('')
  const [endDate, setEndDate] = useState('')

  const handleSubmit = () => {
    create.mutate(
      {
        client_id: clientId,
        type: subType,
        price: price ? Number(price) : undefined,
        lessons_left: subType === 'count' ? Number(lessons) : 0,
        ends_at: endDate,
      },
      { onSuccess: () => onDone() },
    )
  }

  return (
    <Card className="p-5">
      <h4 className="text-sm font-bold text-foreground mb-3">Новый абонемент</h4>
      <div className="flex flex-col gap-3">
        <div className="flex rounded-2xl bg-muted/50 p-1">
          {(['count', 'period'] as const).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setSubType(t)}
              className={`flex-1 rounded-xl py-2 text-sm font-semibold transition-all ${
                subType === t ? 'bg-white shadow-sm text-foreground' : 'text-muted-foreground'
              }`}
            >
              {t === 'count' ? 'По занятиям' : 'Безлимит'}
            </button>
          ))}
        </div>
        {subType === 'count' && (
          <div>
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block">
              Количество занятий
            </label>
            <input
              type="number"
              value={lessons}
              onChange={(e) => setLessons(e.target.value)}
              className="w-full rounded-2xl border border-border/60 bg-white px-4 py-2.5 text-sm"
            />
          </div>
        )}
        <div>
          <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block">
            Стоимость (₽)
          </label>
          <input
            type="number"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            placeholder="5000"
            className="w-full rounded-2xl border border-border/60 bg-white px-4 py-2.5 text-sm"
          />
        </div>
        <div>
          <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block">
            Действует до
          </label>
          <input
            type="date"
            value={endDate}
            onChange={(e) => setEndDate(e.target.value)}
            className="w-full rounded-2xl border border-border/60 bg-white px-4 py-2.5 text-sm"
          />
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={onDone} className="flex-1">
            Отмена
          </Button>
          <Button
            size="sm"
            onClick={handleSubmit}
            disabled={create.isPending || !endDate}
            variant="gradient"
            className="flex-1"
          >
            {create.isPending ? 'Сохранение...' : 'Создать'}
          </Button>
        </div>
      </div>
    </Card>
  )
}

export default function ClientCardScreen({ params }: { params: { id: string } }) {
  const router = useRouter()
  const clientId = parseInt(params.id, 10)
  const { data: clients, isLoading } = useClients()
  const { data: subscriptions, isLoading: subsLoading } = useClientSubscriptions(clientId)
  const deleteSub = useDeleteClientSubscription()
  const [showCreate, setShowCreate] = useState(false)

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
    <div>
      <ScreenHeader title="Клиент" onBack={() => router.back()} />

      <div className="px-5 flex flex-col gap-5">
        {/* Avatar + Name */}
        <div className="flex flex-col items-center pt-2">
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

        <WaveDivider className="text-primary/5 -my-2" />

        {/* Contact */}
<section>
  <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">Контакты</h3>
  <Card className="!p-0 overflow-hidden divide-y divide-border/40">
    {client.phone && (
      <a href={`tel:${client.phone}`} className="flex items-center gap-3 px-4 py-3 hover:bg-muted/30 transition-colors">
        <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 shrink-0">
          <Phone className="h-4 w-4 text-primary" />
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-xs text-muted-foreground">Телефон</p>
          <p className="text-sm font-semibold text-foreground">{client.phone}</p>
        </div>
        <ArrowRight className="h-4 w-4 text-muted-foreground/50 shrink-0" />
      </a>
    )}
    <div className="flex items-center gap-3 px-4 py-3 hover:bg-muted/30 transition-colors">
      <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 shrink-0">
        <MessageCircle className="h-4 w-4 text-primary" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-xs text-muted-foreground">Telegram</p>
        <p className="text-sm font-semibold text-foreground">
          {client.bot_access ? 'Подключён' : 'Не подключён'}
        </p>
      </div>
      <ArrowRight className="h-4 w-4 text-muted-foreground/50 shrink-0" />
    </div>
  </Card>
</section>

        {/* Subscriptions */}
        <section>
          <div className="flex items-center justify-between mb-2 px-1">
            <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground">Абонементы</h3>
            <button
              onClick={() => setShowCreate(true)}
              className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary/10"
            >
              <Plus className="h-4 w-4 text-primary" />
            </button>
          </div>

          {showCreate && (
            <div className="mb-3">
              <CreateSubscriptionForm clientId={clientId} onDone={() => setShowCreate(false)} />
            </div>
          )}

          {subsLoading ? (
            <Spinner />
          ) : subscriptions && subscriptions.length > 0 ? (
            <div className="flex flex-col gap-2">
              {subscriptions.map((sub) => (
                <SubscriptionCard
                  key={sub.id}
                  sub={sub}
                  onDelete={(id) => deleteSub.mutate({ clientId, id })}
                />
              ))}
            </div>
          ) : (
            <Card>
              <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 shrink-0">
                  <CreditCard className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">Абонемент</p>
                  <p className="text-sm font-semibold text-foreground">Не оформлен</p>
                </div>
              </div>
            </Card>
          )}
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
        <Button variant="gradient" size="lg">
          <MessageCircle className="h-5 w-5 mr-2" />
          Написать клиенту
        </Button>
      </div>
    </div>
  )
}
