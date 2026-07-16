'use client'

import { useEffect, useState } from 'react'

export function InstallPrompt() {
  const [deferred, setDeferred] = useState<any>(null)
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const onBefore = (e: Event) => {
      e.preventDefault()
      setDeferred(e)
      setVisible(true)
    }
    window.addEventListener('beforeinstallprompt', onBefore)
    return () => window.removeEventListener('beforeinstallprompt', onBefore)
  }, [])

  const install = async () => {
    if (!deferred) return
    deferred.prompt()
    await deferred.userChoice
    setDeferred(null)
    setVisible(false)
  }

  if (!visible) return null

  return (
    <button
      onClick={install}
      className="fixed bottom-20 left-4 right-4 z-50 rounded-xl bg-tg-button px-4 py-3 text-center text-tg-button-text shadow-lg"
    >
      Установить приложение
    </button>
  )
}
