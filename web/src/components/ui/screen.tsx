import { cn } from '@/lib/utils'
import { ChevronLeft, Loader2, AlertCircle } from 'lucide-react'
export { Card, CardHeader, CardContent } from './card'
export { Button } from './button'

export function Spinner({ label, className }: { label?: string; className?: string }) {
  return (
    <div className={cn('flex flex-col items-center justify-center gap-3 py-12 text-muted-foreground', className)}>
      <Loader2 className="h-8 w-8 animate-spin text-primary" />
      {label && <span className="text-sm font-medium">{label}</span>}
    </div>
  )
}

export function ScreenHeader({
  title,
  subtitle,
  onBack,
  action,
  gradient,
}: {
  title: string
  subtitle?: string
  onBack?: () => void
  action?: React.ReactNode
  gradient?: boolean
}) {
  return (
    <div className="px-5 pb-4 pt-6 flex flex-col gap-1.5">
      <div className="flex items-center justify-between">
        <div className="flex-1">
          {onBack && (
            <button
              onClick={onBack}
              className="mb-1 -ml-1 text-sm font-semibold text-primary inline-flex items-center gap-1 hover:opacity-90 active:scale-95 transition-all"
            >
              <ChevronLeft className="h-4 w-4" />
              <span>Назад</span>
            </button>
          )}
          <h1 className={cn(
            'text-display font-bold tracking-tight',
            gradient ? 'gradient-text' : 'text-foreground',
          )}>
            {title}
          </h1>
          {subtitle && <p className="text-sm font-medium text-muted-foreground mt-0.5">{subtitle}</p>}
        </div>
        {action && <div className="ml-3">{action}</div>}
      </div>
    </div>
  )
}

export function Empty({ text, icon }: { text: string; icon?: React.ReactNode }) {
  return (
    <div className="px-6 py-12 flex flex-col items-center justify-center text-center rounded-3xl border border-dashed border-border bg-muted/30">
      {icon && <div className="mb-3 text-muted-foreground/50">{icon}</div>}
      <p className="text-sm font-medium text-muted-foreground">{text}</p>
    </div>
  )
}

export function ErrorBox({ error }: { error: unknown }) {
  const msg = error instanceof Error ? error.message : 'Произошла ошибка'
  return (
    <div className="mx-4 my-4 rounded-2xl border border-destructive/20 bg-destructive/5 p-4 flex gap-3 text-sm">
      <AlertCircle className="h-5 w-5 shrink-0 text-destructive" />
      <span className="font-medium text-foreground">{msg}</span>
    </div>
  )
}

export function Row({ label, value }: { label: string; value?: string | null }) {
  if (!value) return null
  return (
    <div className="flex justify-between gap-4 py-2.5 border-b border-border/40 last:border-b-0 text-sm">
      <span className="font-medium text-muted-foreground">{label}</span>
      <span className="text-right font-semibold text-foreground">{value}</span>
    </div>
  )
}
