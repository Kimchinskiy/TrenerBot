'use client'

import { useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@/components/auth-provider'
import FloatingNavbar, { type FloatingNavItem } from '@/components/ui/floating-navbar'
import { CalendarDays, User, MoreHorizontal } from 'lucide-react'

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

  const items: FloatingNavItem[] = [
    {
      label: 'Расписание',
      icon: <CalendarDays className="h-5 w-5" />,
      href: '/dashboard/schedule',
    },
    {
      label: 'Профиль',
      icon: <User className="h-5 w-5" />,
      href: '/dashboard/profile',
    },
    {
      label: 'Ещё',
      icon: <MoreHorizontal className="h-5 w-5" />,
      href: '/dashboard/more',
    },
  ]

  const handleNav = (href?: string) => {
    if (href) router.push(href)
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 overflow-y-auto pb-24">{children}</div>
      <FloatingNavbar
        items={items.map((item) => ({
          ...item,
          onClick: () => handleNav(item.href),
        }))}
      />
    </div>
  )
}
