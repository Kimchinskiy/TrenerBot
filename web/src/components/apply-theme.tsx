'use client'

import { useEffect } from 'react'

export function ApplyTheme() {
  useEffect(() => {
    const tw = (window as any).Telegram?.WebApp
    if (tw) {
      tw.ready?.()
      tw.expand?.()
    }
  }, [])
  return null
}
