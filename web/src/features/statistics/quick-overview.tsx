'use client'

import { motion } from 'framer-motion'
import {
  UserPlus,
  Receipt,
  XCircle,
  TrendingUp,
  Users,
  Calendar,
  Clock,
} from 'lucide-react'
import type { QuickOverview as QuickOverviewType } from '@/lib/types'

const rows: {
  key: keyof QuickOverviewType
  icon: React.ElementType
  label: string
  fmt: (v: QuickOverviewType[keyof QuickOverviewType]) => string
}[] = [
  { key: 'new_clients', icon: UserPlus, label: 'Новых клиентов', fmt: (v) => `${v}` },
  { key: 'average_check', icon: Receipt, label: 'Средний чек', fmt: (v) => `${(v as number).toLocaleString('ru-RU')} ₽` },
  { key: 'canceled_count', icon: XCircle, label: 'Отменённых тренировок', fmt: (v) => `${v}` },
  { key: 'avg_attendance', icon: TrendingUp, label: 'Средняя посещаемость', fmt: (v) => `${Math.round(v as number)}%` },
  { key: 'avg_group_size', icon: Users, label: 'Средний размер группы', fmt: (v) => `${(v as number).toFixed(1)}` },
  { key: 'busiest_day', icon: Calendar, label: 'Самый загруженный день', fmt: (v) => `${v || '—'}` },
  { key: 'popular_time', icon: Clock, label: 'Популярное время', fmt: (v) => `${v || '—'}` },
]

export function QuickOverview({ overview }: { overview: QuickOverviewType }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.35, duration: 0.3 }}
      className="rounded-3xl bg-white shadow-card overflow-hidden"
    >
      <p className="text-sm font-bold uppercase tracking-wider text-muted-foreground px-5 pt-5 pb-2">
        Быстрый обзор
      </p>
      <div className="divide-y divide-border/40">
        {rows.map((row) => {
          const value = overview[row.key]
          const Icon = row.icon
          return (
            <div
              key={row.key}
              className="flex items-center gap-3 px-5 py-3.5"
            >
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary/5 shrink-0">
                <Icon className="h-4 w-4 text-primary" />
              </div>
              <span className="flex-1 text-sm text-foreground">{row.label}</span>
              <span className="text-sm font-bold text-foreground">{row.fmt(value)}</span>
            </div>
          )
        })}
      </div>
    </motion.div>
  )
}
