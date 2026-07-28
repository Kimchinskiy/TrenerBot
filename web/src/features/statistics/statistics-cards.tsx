'use client'

import { motion } from 'framer-motion'
import { Dumbbell, Users, Banknote, Activity } from 'lucide-react'
import type { StatisticsResponse } from '@/lib/types'

const cardConfig = [
  {
    key: 'trainings',
    icon: Dumbbell,
    color: 'bg-primary/10 text-primary',
    format: (v: number) => String(v),
    changeFormat: (c: number) => (c >= 0 ? `+${c}` : String(c)),
  },
  {
    key: 'clients',
    icon: Users,
    color: 'bg-blue-500/10 text-blue-500',
    format: (v: number) => String(v),
    changeFormat: (c: number) => (c >= 0 ? `+${c}` : String(c)),
  },
  {
    key: 'income',
    icon: Banknote,
    color: 'bg-emerald-500/10 text-emerald-500',
    format: (v: number) => `${v.toLocaleString('ru-RU')} ₽`,
    changeFormat: (c: number) => `${c >= 0 ? '+' : ''}${c.toFixed(1)}%`,
  },
  {
    key: 'attendance',
    icon: Activity,
    color: 'bg-amber-500/10 text-amber-500',
    format: (v: number) => `${Math.round(v)}%`,
    changeFormat: (c: number) => `${c >= 0 ? '+' : ''}${c.toFixed(1)}%`,
  },
]

function StatCard({
  icon: Icon,
  color,
  label,
  value,
  change,
  format,
  changeFormat,
  index,
}: {
  icon: React.ElementType
  color: string
  label: string
  value: number
  change: number
  format: (v: number) => string
  changeFormat: (c: number) => string
  index: number
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.08, duration: 0.3 }}
      className="flex flex-col gap-2 rounded-3xl bg-card border border-border/50 p-4 shadow-card active:scale-[0.97] transition-transform"
    >
      <div className={`flex h-9 w-9 items-center justify-center rounded-xl ${color}`}>
        <Icon className="h-4 w-4" />
      </div>
      <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
        {label}
      </span>
      <span className="text-display font-bold text-foreground leading-none">
        {format(value)}
      </span>
      <span className={`text-xs font-semibold ${change >= 0 ? 'text-success' : 'text-destructive'}`}>
        {changeFormat(change)}
      </span>
    </motion.div>
  )
}

export function StatisticsCards({ stats }: { stats: StatisticsResponse }) {
  const values = {
    trainings: stats.trainings,
    clients: stats.clients,
    income: stats.income,
    attendance: stats.attendance,
  }

  return (
    <div className="grid grid-cols-2 gap-3">
      {cardConfig.map((cfg, i) => {
        const v = values[cfg.key as keyof typeof values]
        return (
          <StatCard
            key={cfg.key}
            icon={cfg.icon}
            color={cfg.color}
            label={v.label}
            value={v.value}
            change={v.change}
            format={cfg.format}
            changeFormat={cfg.changeFormat}
            index={i}
          />
        )
      })}
    </div>
  )
}
