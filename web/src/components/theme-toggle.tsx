'use client'

import { useTheme, type Theme } from './theme-provider'
import { Sun, Moon, Laptop } from 'lucide-react'

export function ThemeToggle({ className = '' }: { className?: string }) {
  const { resolvedTheme, setTheme } = useTheme()

  const isDark = resolvedTheme === 'dark'

  return (
    <button
      onClick={() => setTheme(isDark ? 'light' : 'dark')}
      className={`flex h-10 w-10 items-center justify-center rounded-2xl bg-muted/80 text-foreground hover:bg-muted active:scale-95 transition-all ${className}`}
      title={isDark ? 'Переключить на светлую тему' : 'Переключить на тёмную тему'}
      aria-label="Переключить тему"
    >
      {isDark ? (
        <Sun className="h-5 w-5 text-amber-400 transition-transform rotate-0 scale-100" />
      ) : (
        <Moon className="h-5 w-5 text-slate-700 transition-transform rotate-0 scale-100" />
      )}
    </button>
  )
}

export function ThemeSelector() {
  const { theme, setTheme } = useTheme()

  const options: { id: Theme; label: string; icon: typeof Sun }[] = [
    { id: 'light', label: 'Светлая', icon: Sun },
    { id: 'dark', label: 'Тёмная', icon: Moon },
    { id: 'system', label: 'Система', icon: Laptop },
  ]

  return (
    <div className="flex rounded-2xl bg-muted p-1 border border-border/40">
      {options.map((opt) => {
        const Icon = opt.icon
        const isActive = theme === opt.id

        return (
          <button
            key={opt.id}
            onClick={() => setTheme(opt.id)}
            className={`flex flex-1 items-center justify-center gap-1.5 rounded-xl py-2 px-3 text-xs font-semibold transition-all ${
              isActive
                ? 'bg-card text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            <Icon className={`h-3.5 w-3.5 ${isActive && opt.id === 'dark' ? 'text-indigo-400' : isActive && opt.id === 'light' ? 'text-amber-500' : ''}`} />
            <span>{opt.label}</span>
          </button>
        )
      })}
    </div>
  )
}
