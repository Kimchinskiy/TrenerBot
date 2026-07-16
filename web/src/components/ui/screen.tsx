import { cn } from '@/lib/utils'
export { Card, CardHeader, CardContent } from './card'
export { Button } from './button'

export function Spinner({ label, className }: { label?: string; className?: string }) {
  return (
    <div className={cn('flex flex-col items-center justify-center gap-3 py-10 text-tg-hint', className)}>
      <div className="h-7 w-7 animate-spin rounded-full border-2 border-tg-hint border-t-transparent" />
      {label && <span className="text-sm">{label}</span>}
    </div>
  )
}

export function ScreenHeader({
  title,
  subtitle,
  onBack,
}: {
  title: string
  subtitle?: string
  onBack?: () => void
}) {
  return (
    <div className="px-4 pb-3 pt-5">
      {onBack && (
        <button onClick={onBack} className="mb-1 text-sm text-tg-link">
          ← Назад
        </button>
      )}
      <h1 className="text-2xl font-bold text-tg-text">{title}</h1>
      {subtitle && <p className="mt-1 text-sm text-tg-hint">{subtitle}</p>}
    </div>
  )
}

export function Empty({ text }: { text: string }) {
  return <div className="px-4 py-10 text-center text-sm text-tg-hint">{text}</div>
}

export function ErrorBox({ error }: { error: unknown }) {
  const msg = error instanceof Error ? error.message : 'Произошла ошибка'
  return <div className="mx-4 my-3 rounded-xl bg-red-500/15 p-3 text-sm text-red-300">{msg}</div>
}

export function Row({ label, value }: { label: string; value?: string | null }) {
  if (!value) return null
  return (
    <div className="flex justify-between gap-4 py-1 text-sm">
      <span className="text-tg-hint">{label}</span>
      <span className="text-right text-tg-text">{value}</span>
    </div>
  )
}
