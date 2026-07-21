'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/components/auth-provider'
import { Spinner } from '@/components/ui/screen'

export default function HomePage() {
  const router = useRouter()
  const { status } = useAuth()

  useEffect(() => {
    if (status === 'authed') router.replace('/dashboard/home')
    else if (status === 'guest') router.replace('/login')
  }, [status, router])

  return (
    <div className="flex h-full items-center justify-center">
      <Spinner label="Загрузка..." />
    </div>
  )
}
