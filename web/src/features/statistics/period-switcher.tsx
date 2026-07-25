'use client'

export function PeriodSwitcher({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const periods = [
    { key: 'week', label: 'Неделя' },
    { key: 'month', label: 'Месяц' },
    { key: 'year', label: 'Год' },
  ]

  return (
    <div className="flex rounded-2xl bg-white shadow-card p-1">
      {periods.map((p) => (
        <button
          key={p.key}
          type="button"
          onClick={() => onChange(p.key)}
          className={`flex-1 rounded-xl py-2.5 text-sm font-semibold transition-all duration-200 ${
            value === p.key
              ? 'bg-primary text-white shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          {p.label}
        </button>
      ))}
    </div>
  )
}
