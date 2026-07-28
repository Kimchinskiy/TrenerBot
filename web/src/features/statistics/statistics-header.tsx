'use client'

import { useRouter } from 'next/navigation'
import { ArrowLeft } from 'lucide-react'

export function StatisticsHeader() {
  const router = useRouter()

  return (
    <div className="flex items-center justify-between pt-2">
      <div className="flex items-center gap-3">
        <button
          onClick={() => router.back()}
          className="flex h-9 w-9 items-center justify-center rounded-xl bg-card border border-border/50 text-foreground shadow-card active:scale-95 transition-transform"
        >
          <ArrowLeft className="h-5 w-5 text-foreground" />
        </button>
        <h1 className="text-title font-bold text-foreground">Статистика</h1>
      </div>
      <div className="text-xs font-medium text-muted-foreground bg-card border border-border/50 rounded-xl px-3 py-2 shadow-card">
        {new Date().toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' })}
      </div>
    </div>
  )
}
