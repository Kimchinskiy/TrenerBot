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
    <div className="fixed bottom-6 left-0 right-0 z-50 flex justify-center px-4">
      <nav className="flex items-center gap-1.5 rounded-full border border-border/60 bg-background/80 p-2 shadow-2xl backdrop-blur-lg transition-all duration-300">
        {items.map((item) => {
          const isActive = item.href ? pathname.startsWith(item.href) : false
          return (
            <button
              key={item.label}
              onClick={item.onClick}
              className={cn(
                'relative flex h-12 w-12 items-center justify-center rounded-full transition-all duration-300 active:scale-95',
                isActive
                  ? 'bg-primary text-primary-foreground shadow-md scale-105'
                  : 'text-muted-foreground hover:bg-muted/40 hover:text-foreground',
              )}
            >
              {item.icon}
              <span className="sr-only">{item.label}</span>
            </button>
          )
        })}
      </nav>
    </div>
  )
}
