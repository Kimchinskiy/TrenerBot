import { ScreenHeader, Card } from '../components/ui'
import type { Role } from '../lib/types'

export type ExtraScreen = 'attendance' | 'wellbeing' | 'contact' | 'debtors' | 'waiting' | 'social'

const LABELS: Record<ExtraScreen, string> = {
  attendance: '✅ Посещаемость',
  wellbeing: '💪 Самочувствие',
  contact: '💬 Связь с тренером',
  debtors: '⚠️ Должники',
  waiting: '📋 Лист ожидания',
  social: '🔗 Соцсети / FAQ',
}

export default function More({
  role,
  onNavigate,
}: {
  role: Role
  onNavigate: (s: ExtraScreen) => void
}) {
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
              hapticsClick()
              onNavigate(s)
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

function hapticsClick() {
  ;(window as any).Telegram?.WebApp?.HapticFeedback?.impactOccurred?.('light')
}
