'use client'

import { useEffect } from 'react'
import { applyTheme } from '@/services/telegram'

export function ApplyTheme() {
  useEffect(() => {
    applyTheme()
  }, [])
  return null
}
