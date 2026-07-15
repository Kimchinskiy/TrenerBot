import type { ReactNode } from 'react'

export function Spinner({ label }: { label?: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-10 text-tg-hint">
      <div className="h-7 w-7 animate-spin rounded-full border-2 border-tg-hint border-t-transparent" />
      {label && <span className="text-sm">{label}</span>}
    </div>
  )
}

export function Card({
  children,
  className = '',
  onClick,
}: {
  children: ReactNode
  className?: string
  onClick?: () => void
}) {
  return (
    <div
      onClick={onClick}
      className={`rounded-2xl bg-tg-secondary p-4 ${onClick ? 'cursor-pointer active:scale-[0.99] transition' : ''} ${className}`}
    >
      {children}
    </div>
  )
}

export function Button({
  children,
  onClick,
  disabled,
  variant = 'primary',
  className = '',
}: {
  children: ReactNode
  onClick?: () => void
  disabled?: boolean
  variant?: 'primary' | 'secondary'
  className?: string
}) {
  const base =
    'w-full rounded-xl px-4 py-3 text-base font-medium transition active:scale-[0.98] disabled:opacity-50'
  const styles =
    variant === 'primary'
      ? 'bg-tg-button text-tg-button-text'
      : 'bg-tg-secondary text-tg-text'
  return (
    <button onClick={onClick} disabled={disabled} className={`${base} ${styles} ${className}`}>
      {children}
    </button>
  )
}

export function ScreenHeader({ title, subtitle, onBack }: { title: string; subtitle?: string; onBack?: () => void }) {
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
  return (
    <div className="mx-4 my-3 rounded-xl bg-red-500/15 p-3 text-sm text-red-300">{msg}</div>
  )
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
