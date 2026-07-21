'use client'

import { Card } from '@/components/ui/card'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import type { ScheduleEntry } from '@/lib/types'
import { Clock, Users, User } from 'lucide-react'

interface LessonCardProps {
  entry: ScheduleEntry
  onClick?: () => void
}

export function LessonCard({ entry, onClick }: LessonCardProps) {
  const isCanceled = entry.status === 'canceled'

  return (
    <Card
      className={`flex items-center gap-3.5 py-3.5 px-4 ${
        isCanceled
          ? 'border-l-4 border-l-destructive bg-destructive/5 opacity-75'
          : 'border-l-4 border-l-primary'
      }`}
      onClick={onClick}
    >
      <Avatar className="h-10 w-10 border border-border/50 shadow-sm shrink-0">
        <AvatarFallback className={`text-xs font-bold uppercase ${
          isCanceled ? 'bg-destructive/10 text-destructive' : 'bg-primary/10 text-primary'
        }`}>
          {entry.group_name ? <Users className="h-4 w-4" /> : <User className="h-4 w-4" />}
        </AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-bold text-foreground truncate">
          {entry.group_name || entry.client_name}
        </p>
        <div className="flex items-center gap-1.5 mt-0.5">
          <Clock className="h-3 w-3 text-muted-foreground" />
          <span className="text-xs text-muted-foreground">{entry.time} · {entry.duration} мин</span>
        </div>
      </div>
      <span className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${
        isCanceled
          ? 'text-destructive bg-destructive/10'
          : 'text-success bg-success-light'
      }`}>
        {isCanceled ? 'Отменено' : 'Активно'}
      </span>
    </Card>
  )
}
