'use client'

import { useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@/components/auth-provider'

function Tab({
  active,
  label,
  href,
}: {
  active: boolean
  label: string
  href: string
}) {
  const router = useRouter()
  return (
    <button
      onClick={() => router.push(href)}
      className={`flex-1 py-3 text-sm font-medium ${active ? 'text-tg-button' : 'text-tg-hint'}`}
    >
      {label}
    </button>
  )
}

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { status } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (status === 'guest') router.replace('/login')
  }, [status, router])

  if (status !== 'authed') {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="h-7 w-7 animate-spin rounded-full border-2 border-tg-hint border-t-transparent" />
      </div>
    )
  }

  const isSchedule = pathname.startsWith('/dashboard/schedule')
  const isProfile = pathname.startsWith('/dashboard/profile')
  const isMore = pathname.startsWith('/dashboard/more')

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 overflow-y-auto">{children}</div>
      <nav className="flex border-t border-tg-secondary bg-tg-bg">
        <Tab active={isSchedule} label="📅 Расписание" href="/dashboard/schedule" />
        <Tab active={isProfile} label="👤 Профиль" href="/dashboard/profile" />
        <Tab active={isMore} label="☰ Ещё" href="/dashboard/more" />
      </nav>
    </div>
  )
}
