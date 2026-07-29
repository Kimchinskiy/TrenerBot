'use client'

import * as React from 'react'
import { usePathname } from 'next/navigation'
import { cn } from '@/lib/utils'

export type FloatingNavItem = {
  label: string
  icon: React.ReactNode
  href?: string
  onClick?: () => void
}

export default function FloatingNavbar({
  items,
}: {
  items: FloatingNavItem[]
}) {
  const pathname = usePathname()

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 flex justify-center px-4 pb-5 pt-2" style={{ paddingBottom: 'calc(1.25rem + var(--safe-bottom))' }}>
      <nav className="glass-strong flex items-center gap-1 rounded-2xl px-2 py-2 shadow-elevated w-full max-w-md">
        {items.map((item) => {
          const isActive = item.href ? pathname.startsWith(item.href) : false
          return (
            <button
              key={item.label}
              onClick={item.onClick}
              className={cn(
                'relative flex flex-1 flex-col items-center gap-0.5 rounded-xl py-2 transition-all duration-200 active:scale-95',
                isActive
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground hover:text-foreground',
              )}
            >
              <div className={cn(
                'flex h-6 w-6 items-center justify-center transition-all duration-200',
                isActive && 'scale-110',
              )}>
                {item.icon}
              </div>
              <span className={cn(
                'text-[10px] font-medium leading-none transition-all duration-200',
                isActive ? 'font-semibold' : 'font-medium',
              )}>
                {item.label}
              </span>
            </button>
          )
        })}
      </nav>
    </div>
  )
}
