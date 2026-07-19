'use client'

import { useRouter } from 'next/navigation'
import { ScreenHeader, Card } from '@/components/ui/screen'
import type { Role } from '@/lib/types'
import { haptics } from '@/services/telegram'
import {
  ClipboardCheck,
  Heart,
  MessageSquare,
  AlertTriangle,
  ListOrdered,
  Globe,
  Send,
  ChevronRight,
} from 'lucide-react'

export type ExtraScreen = 'attendance' | 'wellbeing' | 'contact' | 'debtors' | 'waiting' | 'social' | 'notifications'

const ICONS: Record<ExtraScreen, React.ReactNode> = {
  attendance: <ClipboardCheck className="h-5 w-5 text-emerald-400" />,
  wellbeing: <Heart className="h-5 w-5 text-rose-450" />,
  contact: <MessageSquare className="h-5 w-5 text-blue-400" />,
  debtors: <AlertTriangle className="h-5 w-5 text-amber-500" />,
  waiting: <ListOrdered className="h-5 w-5 text-indigo-400" />,
  social: <Globe className="h-5 w-5 text-cyan-400" />,
  notifications: <Send className="h-5 w-5 text-purple-400" />,
}

const LABELS: Record<ExtraScreen, string> = {
  attendance: 'Посещаемость',
  wellbeing: 'Самочувствие',
  contact: 'Связь с тренером',
  debtors: 'Должники',
  waiting: 'Лист ожидания',
  social: 'Соцсети / FAQ',
  notifications: 'Оповестить клиентов',
}

export default function More({ role }: { role: Role }) {
  const router = useRouter()

  const items: ExtraScreen[] =
    role === 'client'
      ? ['wellbeing', 'contact', 'social']
      : role === 'parent'
        ? ['contact', 'social']
        : role === 'coach'
          ? ['attendance', 'debtors', 'waiting', 'contact', 'social', 'notifications']
        : role === 'admin'
          ? ['attendance', 'debtors', 'waiting', 'social', 'notifications']
          : ['attendance', 'debtors', 'waiting', 'social']

  return (
    <div>
      <ScreenHeader title="Ещё" subtitle="Дополнительные функции CRM" />
      <div className="flex flex-col gap-3 px-4 pb-24">
        {items.map((s) => (
          <Card
            key={s}
            className="flex items-center justify-between py-4 px-5 hover:bg-muted/30 hover:border-foreground/10 active:scale-[0.99] transition-all duration-200"
            onClick={() => {
              haptics.light()
              router.push(`/dashboard/more/${s}`)
            }}
          >
            <div className="flex items-center gap-3.5">
              <div className="p-2 rounded-xl bg-card border border-border shadow-sm shrink-0">
                {ICONS[s]}
              </div>
              <span className="text-base font-bold text-foreground">{LABELS[s]}</span>
            </div>
            <ChevronRight className="h-5 w-5 text-muted-foreground/80 shrink-0" />
          </Card>
        ))}
      </div>
    </div>
  )
}
