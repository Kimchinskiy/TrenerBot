'use client'

import { useEffect, useRef } from 'react'

const BOT_USERNAME = process.env.NEXT_PUBLIC_TELEGRAM_BOT || ''

export function TelegramLoginButton({
  onAuth,
}: {
  onAuth: (fields: Record<string, string>) => void
}) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!BOT_USERNAME || !ref.current) return
    ;(window as any).onTelegramAuth = (user: Record<string, unknown>) => {
      const fields: Record<string, string> = {}
      Object.entries(user).forEach(([k, v]) => (fields[k] = String(v)))
      onAuth(fields)
    }
    const script = document.createElement('script')
    script.src = 'https://telegram.org/js/telegram-widget.js?22'
    script.async = true
    script.setAttribute('data-telegram-login', BOT_USERNAME)
    script.setAttribute('data-size', 'large')
    script.setAttribute('data-radius', '12')
    script.setAttribute('data-onauth', 'onTelegramAuth(user)')
    script.setAttribute('data-request-access', 'write')
    ref.current.appendChild(script)
    return () => {
      if (ref.current) ref.current.innerHTML = ''
    }
  }, [onAuth])

  if (!BOT_USERNAME) return null
  return <div ref={ref} className="flex justify-center pt-2" />
}
