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
      <div className="flex gap-2.5 px-4 pb-4">
        {PERIODS.map((p) => (
          <button
            key={p.days}
            onClick={() => setDays(p.days)}
            className={`flex-1 rounded-xl py-2 text-sm font-semibold border transition-all duration-200 active:scale-95 ${
              days === p.days
                ? 'bg-primary text-primary-foreground border-primary shadow-sm'
                : 'bg-card border-border text-foreground hover:bg-muted/40'
            }`}
          >
            {p.label}
          </button>
        ))}
      </div>
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}
      {!isLoading && data && data.length === 0 && <Empty text="Должников нет 🎉" />}
      <div className="flex flex-col gap-3 px-4 pb-24">
        {(data || []).map((d) => (
          <Card key={d.client_id} className="border-border/80 shadow-sm relative overflow-hidden">
            <div className="absolute top-0 right-0 h-16 w-16 bg-primary/5 rounded-full blur-2xl" />
            <h2 className="text-base font-bold text-foreground mb-2">{d.name}</h2>
            <div className="flex flex-col gap-1.5 pt-1.5">
              <Row label="Пропущено" value={`${d.missed_count} тренировок`} />
              <Row label="Телефон" value={d.phone} />
            </div>
          </Card>
        ))}
      </div>
    </div>
  )
}
