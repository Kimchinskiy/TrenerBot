'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useDebtors } from '@/lib/hooks'
import { Card, ScreenHeader, Spinner, Empty, ErrorBox, Row } from '@/components/ui/screen'

const PERIODS = [
  { days: 7, label: '7 дней' },
  { days: 30, label: '30 дней' },
  { days: 90, label: '90 дней' },
]

export default function DebtorsScreen() {
  const router = useRouter()
  const [days, setDays] = useState(30)
  const { data, isLoading, error } = useDebtors(days)

  return (
    <div>
      <ScreenHeader title="Должники" subtitle="Пропущенные тренировки за период" onBack={() => router.back()} />
      <div className="flex gap-2 px-4 pb-3">
        {PERIODS.map((p) => (
          <button
            key={p.days}
            onClick={() => setDays(p.days)}
            className={`flex-1 rounded-xl py-2 text-sm ${
              days === p.days ? 'bg-tg-button text-tg-button-text' : 'bg-tg-secondary text-tg-text'
            }`}
          >
            {p.label}
          </button>
        ))}
      </div>
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}
      {!isLoading && data && data.length === 0 && <Empty text="Должников нет 🎉" />}
      <div className="flex flex-col gap-2 px-4 pb-24">
        {(data || []).map((d) => (
          <Card key={d.client_id} className="mb-2">
            <div className="font-semibold">{d.name}</div>
            <Row label="Пропущено" value={`${d.missed_count} тр.`} />
            <Row label="Телефон" value={d.phone} />
          </Card>
        ))}
      </div>
    </div>
  )
}
