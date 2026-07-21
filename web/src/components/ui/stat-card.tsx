import { cn } from '@/lib/utils'
import type { LucideIcon } from 'lucide-react'

interface StatCardProps {
  label: string
  value: string | number
  change?: string
  icon: LucideIcon
  variant?: 'default' | 'primary' | 'success' | 'warning'
}

const variantStyles = {
  default: 'bg-muted/50 text-muted-foreground',
  primary: 'bg-primary/10 text-primary',
  success: 'bg-success-light text-success',
  warning: 'bg-warning-light text-warning',
}

export function StatCard({ label, value, change, icon: Icon, variant = 'default' }: StatCardProps) {
  return (
    <div className="flex flex-col gap-2 rounded-2xl bg-white p-4 shadow-card border border-border/30">
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{label}</span>
        <div className={cn('flex h-7 w-7 items-center justify-center rounded-lg', variantStyles[variant])}>
          <Icon className="h-3.5 w-3.5" />
        </div>
      </div>
      <div className="flex items-end gap-1.5">
        <span className="text-2xl font-bold tracking-tight text-foreground">{value}</span>
        {change && (
          <span className="text-xs font-semibold text-success mb-1">{change}</span>
        )}
      </div>
    </div>
  )
}
