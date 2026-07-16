'use client'

import { useRouter } from 'next/navigation'
import { ScreenHeader, Card } from '@/components/ui/screen'
import type { Role } from '@/lib/types'
import { haptics } from '@/services/telegram'

export type ExtraScreen = 'attendance' | 'wellbeing' | 'contact' | 'debtors' | 'waiting' | 'social'

const LABELS: Record<ExtraScreen, string> = {
  attendance: '✅ Посещаемость',
  wellbeing: '💪 Самочувствие',
  contact: '💬 Связь с тренером',
  debtors: '⚠️ Должники',
  waiting: '📋 Лист ожидания',
  social: '🔗 Соцсети / FAQ',
}

export default function More({ role }: { role: Role }) {
  const router = useRouter()

  const items: ExtraScreen[] =
    role === 'client'
      ? ['wellbeing', 'contact', 'social']
      : role === 'parent'
        ? ['contact', 'social']
        : role === 'coach'
          ? ['attendance', 'debtors', 'waiting', 'contact', 'social']
          : ['attendance', 'debtors', 'waiting', 'social']

  return (
    <div>
      <ScreenHeader title="Ещё" />
      <div className="flex flex-col gap-2 px-4 pb-24">
        {items.map((s) => (
          <Card
            key={s}
            className="flex items-center justify-between"
            onClick={() => {
              haptics.light()
              router.push(`/dashboard/more/${s}`)
            }}
          >
            <span className="text-base font-medium">{LABELS[s]}</span>
            <span className="text-tg-link">→</span>
          </Card>
        ))}
      </div>
    </div>
  )
}
