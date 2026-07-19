'use client'

import { Card } from '@/components/ui/card'
import type { ScheduleEntry } from '@/lib/types'
import { Clock, User } from 'lucide-react'

interface LessonCardProps {
  entry: ScheduleEntry
  onClick?: () => void
}

export function LessonCard({ entry, onClick }: LessonCardProps) {
  const isCanceled = entry.status === 'canceled'

  return (
    <Card
      className={`mb-3 flex items-center justify-between border-l-4 transition-all duration-200 ${
        isCanceled
          ? 'border-l-destructive bg-destructive/5 opacity-75'
          : 'border-l-primary bg-card/50 hover:bg-muted/30'
      }`}
      onClick={onClick}
    >
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
          <Clock className="h-3.5 w-3.5 text-primary" />
          <span>{entry.time}</span>
          <span className="text-border">·</span>
          <span>{entry.duration} мин</span>
        </div>
        <div className="flex items-center gap-1.5 text-base font-bold text-foreground">
          <User className="h-4 w-4 text-muted-foreground shrink-0" />
          <span className="truncate">{entry.client_name}</span>
        </div>
      </div>

      <div>
        <span
          className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-[11px] font-bold select-none border tracking-wide uppercase transition-colors ${
            isCanceled
              ? 'border-destructive/30 bg-destructive/15 text-destructive-foreground'
              : 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400'
          }`}
        >
          {isCanceled ? 'Отменено' : 'Активно'}
        </span>
      </div>
    </Card>
  )
}
