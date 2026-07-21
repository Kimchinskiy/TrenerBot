'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export default function DashboardRootPage() {
  const router = useRouter()

  useEffect(() => {
    router.replace('/dashboard/home')
  }, [router])

  return null
}