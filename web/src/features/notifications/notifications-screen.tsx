'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { ScreenHeader, Card, Empty } from '@/components/ui/screen'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useClients } from '@/lib/hooks'
import { api } from '@/lib/api'
import { haptics } from '@/services/telegram'
import { Send, Bell, MessageCircle, AlertTriangle, CheckCircle2 } from 'lucide-react'

type Tab = 'all' | 'reminders' | 'system'

export default function NotificationsScreen() {
  const router = useRouter()
  const { data: clients } = useClients()
  const [tab, setTab] = useState<Tab>('all')
  const [title, setTitle] = useState('')
  const [text, setText] = useState('')
  const [sending, setSending] = useState(false)
  const [sent, setSent] = useState(false)
  const [error, setError] = useState('')

  const handleSend = async () => {
    if (!title.trim() || !text.trim()) return
    setSending(true)
    setError('')
    try {
      await api.notificationsSend({
        filter: 'all',
        title: title.trim(),
        text: text.trim(),
      })
      setSent(true)
      haptics.success()
      setTitle('')
      setText('')
      setTimeout(() => setSent(false), 3000)
    } catch {
      setError('Ошибка отправки')
      haptics.error()
    } finally {
      setSending(false)
    }
  }

  return (
    <div>
      <ScreenHeader title="Оповещения" onBack={() => router.back()} />

      <div className="px-5 flex flex-col gap-5">
        {/* Tabs */}
        <div className="flex gap-2">
          {([
            { key: 'all', label: 'Все', icon: Bell },
            { key: 'reminders', label: 'Напоминания', icon: MessageCircle },
            { key: 'system', label: 'Системные', icon: AlertTriangle },
          ] as const).map(({ key, label, icon: Icon }) => (
            <button
              key={key}
              onClick={() => setTab(key)}
              className={`flex items-center gap-1.5 rounded-xl px-3 py-2 text-xs font-semibold transition-all duration-200 ${
                tab === key
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              }`}
            >
              <Icon className="h-3.5 w-3.5" />
              {label}
            </button>
          ))}
        </div>

        {/* Send form */}
        <section>
          <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">
            Отправить оповещение
          </h3>
          <Card className="flex flex-col gap-3">
            <Input
              placeholder="Заголовок"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
            <Input
              placeholder="Текст сообщения"
              value={text}
              onChange={(e) => setText(e.target.value)}
              className="h-20"
            />

            {sent && (
              <div className="flex items-center gap-2 text-sm font-semibold text-success">
                <CheckCircle2 className="h-4 w-4" />
                Оповещение отправлено
              </div>
            )}
            {error && (
              <p className="text-sm font-semibold text-destructive">{error}</p>
            )}

            <Button
              onClick={handleSend}
              disabled={!title.trim() || !text.trim() || sending}
              variant="gradient"
            >
              <Send className="h-4 w-4 mr-2" />
              {sending ? 'Отправка...' : `Отправить ${clients?.length || 0} клиентам`}
            </Button>
          </Card>
        </section>

        {/* History placeholder */}
        <section>
          <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">
            История отправок
          </h3>
          <Empty text="Пока нет отправленных оповещений" />
        </section>
      </div>
    </div>
  )
}
