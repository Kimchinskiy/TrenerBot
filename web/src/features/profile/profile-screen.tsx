'use client'

import { useMe } from '@/lib/hooks'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox, Row } from '@/components/ui/screen'
import type { Client } from '@/lib/types'

function ClientCard({ c }: { c: Client }) {
  return (
    <Card className="mb-2">
      <div className="text-lg font-semibold">{c.full_name}</div>
      <Row label="Статус" value={c.status === 'active' ? 'Активен' : c.status} />
      <Row label="Телефон" value={c.phone} />
      <Row label="Возраст" value={c.age ? String(c.age) : null} />
      <Row label="Мед. ограничения" value={c.medical_limits} />
      {c.subscription_ends_at && <Row label="Подписка до" value={c.subscription_ends_at} />}
    </Card>
  )
}

export default function Profile() {
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
            <div className="mt-3">
              <div className="mb-2 px-1 text-sm font-semibold text-tg-hint">Дети</div>
              {data.children && data.children.length > 0 ? (
                data.children.map((c) => <ClientCard key={c.id} c={c} />)
              ) : (
                <Empty text="Нет привязанных детей" />
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
