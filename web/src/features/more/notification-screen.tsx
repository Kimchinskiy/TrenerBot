'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import { ScreenHeader, Card, Spinner, ErrorBox } from '@/components/ui/screen'
import { Button, Input, Label } from '@/components/ui'
import { haptics } from '@/services/telegram'
import type { Recipient } from '@/lib/types'

type RecipientFilter = 'all' | 'today' | 'tomorrow' | 'manual'

const FILTER_LABELS: Record<RecipientFilter, string> = {
  all: 'Всем клиентам',
  today: 'Клиентам сегодняшних тренировок',
  tomorrow: 'Клиентам завтрашних тренировок',
  manual: 'Выбранным вручную',
}

export default function NotificationScreen() {
  const router = useRouter()
  const [filter, setFilter] = useState<RecipientFilter>('all')
  const [title, setTitle] = useState('')
  const [text, setText] = useState('')
  const [preview, setPreview] = useState<{ total: number; recipients: Recipient[] } | null>(null)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(false)
  const [sending, setSending] = useState(false)
  const [result, setResult] = useState<{ total: number; enqueued: number; skipped: number; errors: number } | null>(null)
  const [error, setError] = useState('')
  const [confirming, setConfirming] = useState(false)

  const loadPreview = useCallback(async () => {
    setError('')
    setLoading(true)
    setPreview(null)
    setResult(null)
    try {
      const body: any = { filter }
      if (filter === 'manual') body.client_ids = [...selectedIds]
      const res = await api.notificationsPreview(body)
      setPreview(res)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка загрузки')
    } finally {
      setLoading(false)
    }
  }, [filter, selectedIds])

  const handleSend = async () => {
    if (!text.trim()) { setError('Введите текст сообщения'); return }
    if (filter === 'manual' && selectedIds.size === 0) { setError('Выберите получателей'); return }
    setConfirming(true)
  }

  const confirmSend = async () => {
    setError('')
    setSending(true)
    setConfirming(false)
    try {
      const body: any = { filter, title, text }
      if (filter === 'manual') body.client_ids = [...selectedIds]
      const res = await api.notificationsSend(body)
      setResult(res)
      haptics.success()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка отправки')
    } finally {
      setSending(false)
    }
  }

  const toggleClient = (id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const selectAll = () => {
    if (preview) {
      setSelectedIds(new Set(preview.recipients.map((r) => r.client_id)))
    }
  }

  const deselectAll = () => setSelectedIds(new Set())

  const filteredRecipients = preview?.recipients.filter((r) =>
    search ? r.full_name.toLowerCase().includes(search.toLowerCase()) : true,
  )

  return (
    <div>
      <ScreenHeader title="Оповестить клиентов" onBack={() => router.back()} />
      <div className="flex flex-col gap-3 px-4 pb-24">
        {error && <ErrorBox error={error} />}

        {/* Step 1: Choose filter */}
        {!preview && !result && !sending && !confirming && (
          <>
            <div className="flex flex-col gap-3">
              <div className="mb-1 px-1 text-sm font-bold tracking-wider text-muted-foreground uppercase">Кому отправляем?</div>
              {(['all', 'today', 'tomorrow', 'manual'] as RecipientFilter[]).map((f) => (
                <Card
                  key={f}
                  className={`flex items-center justify-between py-4 px-5 border transition-all duration-200 cursor-pointer ${
                    filter === f ? 'border-primary bg-primary/5 shadow-sm' : 'border-border bg-card hover:bg-muted/10'
                  }`}
                  onClick={() => { haptics.light(); setFilter(f); setSelectedIds(new Set()) }}
                >
                  <span className="text-base font-bold text-foreground">{FILTER_LABELS[f]}</span>
                  <span className={`text-base font-extrabold ${filter === f ? 'text-primary' : 'text-muted-foreground/60'}`}>
                    {filter === f ? '●' : '○'}
                  </span>
                </Card>
              ))}
            </div>

            {/* Manual client picker */}
            {filter === 'manual' && !preview && (
              <Button onClick={loadPreview} disabled={loading} className="font-bold mt-2">
                {loading ? 'Загрузка...' : 'Выбрать клиентов'}
              </Button>
            )}

            {/* Non-manual: load preview directly */}
            {filter !== 'manual' && (
              <Button onClick={loadPreview} disabled={loading} className="font-bold mt-2">
                {loading ? 'Поиск получателей...' : 'Далее'}
              </Button>
            )}
          </>
        )}

        {/* Step 1.5: Manual client picker */}
        {filter === 'manual' && preview && (
          <>
            <div className="flex items-center justify-between">
              <Label className="mb-0">Выберите клиентов</Label>
              <div className="flex gap-2 text-xs font-semibold">
                <button onClick={selectAll} className="text-primary hover:underline">Выбрать всех</button>
                <span className="text-muted-foreground/40">|</span>
                <button onClick={deselectAll} className="text-primary hover:underline">Снять выбор</button>
              </div>
            </div>
            <Input
              placeholder="Поиск..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            <div className="flex flex-col gap-2 max-h-64 overflow-y-auto pr-1">
              {filteredRecipients?.map((r) => {
                const isSelected = selectedIds.has(r.client_id)
                return (
                  <Card
                    key={r.client_id}
                    className={`flex items-center justify-between py-3 px-4 border transition-all duration-200 cursor-pointer ${
                      isSelected ? 'border-primary bg-primary/5 shadow-sm' : 'border-border bg-card'
                    }`}
                    onClick={() => toggleClient(r.client_id)}
                  >
                    <span className="text-sm font-bold text-foreground">{r.full_name}</span>
                    <span className="text-sm font-extrabold text-primary">{isSelected ? '✓' : ''}</span>
                  </Card>
                )
              })}
            </div>
            <div className="text-xs font-semibold text-muted-foreground px-1">
              Выбрано: {selectedIds.size} из {preview.recipients.length}
            </div>
            <Button onClick={() => { setPreview(null); loadPreview() }} disabled={loading || selectedIds.size === 0} className="font-bold mt-2">
              {loading ? 'Загрузка...' : 'Далее'}
            </Button>
          </>
        )}

        {/* Step 2: Compose message */}
        {preview && filter !== 'manual' && !result && !confirming && (
          <>
            <Card className="border-border bg-card shadow-sm p-4 relative overflow-hidden">
               <div className="absolute top-0 right-0 h-16 w-16 bg-primary/5 rounded-full blur-2xl" />
              <div className="text-sm font-bold text-muted-foreground mb-2">Получателей: {preview.total}</div>
              <div className="flex flex-col gap-1.5 max-h-32 overflow-y-auto pr-1">
                {preview.recipients.map((r) => (
                  <div key={r.client_id} className="text-sm font-semibold text-foreground">• {r.full_name}</div>
                ))}
              </div>
            </Card>
            <Input placeholder="Заголовок (необязательно)" value={title} onChange={(e) => setTitle(e.target.value)} />
            <div>
              <textarea
                className="w-full rounded-xl border border-input bg-transparent p-3 text-base ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 resize-none shadow-sm"
                rows={5}
                placeholder="Текст сообщения"
                value={text}
                onChange={(e) => setText(e.target.value)}
              />
            </div>
            <Button onClick={handleSend} disabled={!text.trim() || loading} className="font-bold mt-2">
              Отправить
            </Button>
          </>
        )}

        {/* Step 2.5: manual compose */}
        {preview && filter === 'manual' && !result && !confirming && (
          <>
            <Card className="border-border bg-card shadow-sm p-4">
              <div className="text-sm font-bold text-muted-foreground">Получателей: {selectedIds.size} из {preview.total}</div>
            </Card>
            <Input placeholder="Заголовок (необязательно)" value={title} onChange={(e) => setTitle(e.target.value)} />
            <div>
              <textarea
                className="w-full rounded-xl border border-input bg-transparent p-3 text-base ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 resize-none shadow-sm"
                rows={5}
                placeholder="Текст сообщения"
                value={text}
                onChange={(e) => setText(e.target.value)}
              />
            </div>
            <Button onClick={handleSend} disabled={!text.trim() || loading} className="font-bold mt-2">
              Отправить
            </Button>
          </>
        )}

        {/* Confirm dialog */}
        {confirming && (
          <Card className="p-6 text-center border-border bg-card shadow-xl max-w-md mx-auto">
            <div className="text-lg font-bold mb-2 text-foreground">Отправить сообщение выбранным получателям?</div>
            <div className="text-sm font-semibold text-muted-foreground mb-5">
              Количество получателей: {filter === 'manual' ? selectedIds.size : preview?.total || 0}
            </div>
            <div className="flex gap-3 justify-center">
              <Button variant="outline" size="sm" onClick={() => setConfirming(false)}>Отмена</Button>
              <Button size="sm" onClick={confirmSend} disabled={sending} className="font-bold">
                {sending ? 'Отправка...' : 'Подтвердить'}
              </Button>
            </div>
          </Card>
        )}

        {/* Sending spinner */}
        {sending && <Spinner label="Отправка оповещений..." />}

        {/* Step 3: Result */}
        {result && (
          <Card className="p-6 border-border bg-card shadow-xl max-w-md mx-auto">
            <div className="text-lg font-bold mb-4 text-center text-foreground">Результат отправки</div>
            <div className="flex flex-col gap-2.5">
              <div className="flex justify-between py-1.5 border-b border-border/40 text-sm">
                <span className="font-semibold text-muted-foreground">Всего клиентов:</span>
                <span className="font-bold text-foreground">{result.total}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-border/40 text-sm">
                <span className="font-semibold text-muted-foreground">Отправлено:</span>
                <span className="font-bold text-emerald-400">{result.enqueued}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-border/40 text-sm">
                <span className="font-semibold text-muted-foreground">Пропущено (нет TG):</span>
                <span className="font-bold text-amber-500">{result.skipped}</span>
              </div>
              <div className="flex justify-between py-1.5 last:border-b-0 text-sm">
                <span className="font-semibold text-muted-foreground">Ошибок:</span>
                <span className={`font-bold ${result.errors > 0 ? 'text-destructive' : 'text-foreground'}`}>{result.errors}</span>
              </div>
            </div>
            <div className="mt-6 flex justify-center">
              <Button onClick={() => { setPreview(null); setResult(null); setSelectedIds(new Set()); setTitle(''); setText(''); setFilter('all') }} className="font-bold max-w-xs shadow-md">
                Готово
              </Button>
            </div>
          </Card>
        )}
      </div>
    </div>
  )
}
