'use client'

import { useAuth } from '@/components/auth-provider'
import More from '@/features/more/more-screen'

export default function Page() {
  const { role } = useAuth()
  return <More role={role ?? 'client'} />
}
