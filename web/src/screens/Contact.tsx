import { useState } from 'react'
import { useMe, useMessageCoach } from '../lib/hooks'
import { ScreenHeader, Button, Spinner, ErrorBox } from '../components/ui'
import { haptics } from '../lib/theme'

export default function Contact({ onBack }: { onBack: () => void }) {
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
      <ScreenHeader title="Связь с тренером" subtitle="Сообщение придёт всем тренерам" onBack={onBack} />
      {me.isLoading && <Spinner label="Загрузка..." />}
      {me.isError && <ErrorBox error={me.error} />}
      <div className="px-4 pb-24">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Напишите сообщение тренеру..."
          className="w-full rounded-xl bg-tg-secondary p-3 text-tg-text outline-none"
          rows={5}
        />
        <div className="mt-4">
          <Button onClick={onSend} disabled={!text.trim() || send.isPending}>
            {send.isPending ? 'Отправка...' : 'Отправить тренеру'}
          </Button>
        </div>
        {send.isSuccess && <div className="mt-3 text-center text-sm text-green-400">Отправлено ✅</div>}
      </div>
    </div>
  )
}
