'use client'

import { useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@/components/auth-provider'
import FloatingNavbar, { type FloatingNavItem } from '@/components/ui/floating-navbar'
import { Home, CalendarDays, Users, CheckCircle, User, Layers } from 'lucide-react'
import { useMe } from '@/lib/hooks'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { status } = useAuth()
  const router = useRouter()
  const pathname = usePathname()
  const { data: me } = useMe()

  useEffect(() => {
    if (status === 'guest') router.replace('/login')
  }, [status, router])

  if (status !== 'authed') {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="h-7 w-7 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    )
  }

  const items: FloatingNavItem[] = [
    {
      label: 'Главная',
      icon: <Home className="h-5 w-5" />,
      href: '/dashboard/home',
    },
    {
      label: 'Расписание',
      icon: <CalendarDays className="h-5 w-5" />,
      href: '/dashboard/schedule',
    },
    ...(me?.role === 'coach' || me?.role === 'admin'
      ? [
          {
            label: 'Группы',
            icon: <Layers className="h-5 w-5" />,
            href: '/dashboard/groups',
          },
        ]
      : []),
    ...(me?.role === 'coach' || me?.role === 'admin'
      ? [
          {
            label: 'Клиенты',
            icon: <Users className="h-5 w-5" />,
            href: '/dashboard/clients',
          },
        ]
      : []),
    {
      label: 'Посещаемость',
      icon: <CheckCircle className="h-5 w-5" />,
      href: '/dashboard/attendance',
    },
    {
      label: 'Профиль',
      icon: <User className="h-5 w-5" />,
      href: '/dashboard/profile',
    },
  ]

  const handleNav = (href?: string) => {
    if (href) router.push(href)
  }

  return (
    <div className="flex h-full flex-col bg-background">
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
