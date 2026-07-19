'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useMe, useMessageCoach } from '@/lib/hooks'
import { ScreenHeader, Button, Spinner, ErrorBox } from '@/components/ui/screen'
import { haptics } from '@/services/telegram'

export default function ContactScreen() {
  const router = useRouter()
  const me = useMe()
  const [text, setText] = useState('')
  const send = useMessageCoach()

  const from = me.data?.client?.full_name || 'Клиент'

  const onSend = async () => {
    if (!text.trim()) return
    try {
      await send.mutateAsync({ from, text: text.trim() })
      haptics.success()
      setText('')
    } catch {
      haptics.error()
    }
  }

  return (
    <div>
      <ScreenHeader title="Связь с тренером" subtitle="Сообщение придёт всем тренерам" onBack={() => router.back()} />
      {me.isLoading && <Spinner label="Загрузка..." />}
      {me.isError && <ErrorBox error={me.error} />}
      <div className="px-4 pb-24">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Напишите сообщение тренеру..."
          className="w-full rounded-xl border border-input bg-transparent p-3 text-base ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 resize-none shadow-sm"
          rows={5}
        />
        <div className="mt-4">
          <Button onClick={onSend} disabled={!text.trim() || send.isPending} className="font-bold">
            {send.isPending ? 'Отправка...' : 'Отправить тренеру'}
          </Button>
        </div>
        {send.isSuccess && <div className="mt-3 text-center text-sm font-semibold text-emerald-400">Отправлено ✅</div>}
      </div>
    </div>
  )
}
