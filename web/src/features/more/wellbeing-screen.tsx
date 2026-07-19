'use client'

import { useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useMe, useSchedule, useSubmitWellbeing, useWellbeingHistory } from '@/lib/hooks'
import { weekRange, prettyDate } from '@/lib/dates'
import { Card, ScreenHeader, Spinner, Button, Empty, ErrorBox } from '@/components/ui/screen'
import { haptics } from '@/services/telegram'

const RATINGS = [
  { v: 1, e: '😣', l: 'Очень тяжело' },
  { v: 2, e: '😓', l: 'Тяжело' },
  { v: 3, e: '🙂', l: 'Нормально' },
  { v: 4, e: '😊', l: 'Хорошо' },
  { v: 5, e: '🔥', l: 'Отлично' },
]

export default function WellbeingScreen() {
  const router = useRouter()
  const me = useMe()
  const { from, to } = useMemo(() => weekRange(), [])
  const { data: lessons, isLoading } = useSchedule(from, to)
  const clientId = me.data?.client?.id ?? 0
  const history = useWellbeingHistory(clientId)
  const submit = useSubmitWellbeing()

  const [lessonId, setLessonId] = useState<number | null>(null)
  const [rating, setRating] = useState<number>(5)
  const [note, setNote] = useState('')

  const clientLessons = (lessons || []).filter((l) => l.status !== 'canceled')

  const onSend = async () => {
    if (!lessonId) return
    try {
      await submit.mutateAsync({ lessonId, wellbeing: rating, note })
      haptics.success()
      setNote('')
    } catch {
      haptics.error()
    }
  }

  return (
    <div>
      <ScreenHeader title="Самочувствие" subtitle="Оцените тренировку" onBack={() => router.back()} />
      {isLoading && <Spinner label="Загрузка..." />}
      {me.isError && <ErrorBox error={me.error} />}

      <div className="px-4 pb-6">
        <div className="mb-2 text-sm text-tg-hint">Тренировка</div>
        <div className="flex flex-col gap-2">
          {clientLessons.map((l) => (
            <Card
              key={l.id}
              className={`flex items-center justify-between ${lessonId === l.id ? 'ring-2 ring-tg-button' : ''}`}
              onClick={() => setLessonId(l.id)}
            >
              <div>
                <div className="font-semibold">
                  {prettyDate(l.date)} · {l.time}
                </div>
                {l.location && <div className="text-sm text-tg-hint">📍 {l.location}</div>}
              </div>
            </Card>
          ))}
          {clientLessons.length === 0 && <Empty text="Нет тренировок для оценки" />}
        </div>

        <div className="mb-3 mt-5 text-sm font-bold tracking-wider text-muted-foreground uppercase">Оценка самочувствия</div>
        <div className="flex justify-between gap-2.5">
          {RATINGS.map((r) => (
            <button
              key={r.v}
              onClick={() => setRating(r.v)}
              className={`flex flex-1 flex-col items-center rounded-2xl py-3 border.5 transition-all duration-200 active:scale-95 ${
                rating === r.v
                  ? 'bg-primary text-primary-foreground border-primary shadow-md scale-105 font-bold'
                  : 'bg-card border-border text-foreground hover:bg-muted/40 font-medium'
              }`}
            >
              <span className="text-2xl mb-1">{r.e}</span>
              <span className="text-[10px] tracking-tight">{r.l}</span>
            </button>
          ))}
        </div>

        <textarea
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="Напишите комментарий (например: болели мышцы ног, кружилась голова...)"
          className="mt-4 w-full rounded-xl border border-input bg-transparent p-3 text-base ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 resize-none shadow-sm"
          rows={3}
        />

        <div className="mt-5">
          <Button onClick={onSend} disabled={!lessonId || submit.isPending} className="font-bold flex items-center justify-center gap-2">
            {submit.isPending ? 'Отправка...' : 'Отправить оценку'}
          </Button>
        </div>
      </div>

      <div className="px-4 pb-24">
        <div className="mb-3 px-1 text-sm font-bold tracking-wider text-muted-foreground uppercase">История оценок</div>
        {history.data && history.data.length > 0 ? (
          history.data.map((h) => (
            <Card key={h.id} className="mb-3 border-border/80 shadow-sm relative overflow-hidden">
              <div className="absolute top-0 right-0 h-16 w-16 bg-primary/5 rounded-full blur-2xl" />
              <div className="flex items-center justify-between">
                <span className="text-base font-bold text-foreground">
                  {h.date} · {h.time}
                </span>
                <span className="text-sm font-medium text-muted-foreground text-right">{h.note}</span>
              </div>
            </Card>
          ))
        ) : (
          <Empty text="Пока нет оценок" />
        )}
      </div>
    </div>
  )
}
