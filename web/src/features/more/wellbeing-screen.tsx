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

        <div className="mb-2 mt-4 text-sm text-tg-hint">Оценка</div>
        <div className="flex justify-between gap-2">
          {RATINGS.map((r) => (
            <button
              key={r.v}
              onClick={() => setRating(r.v)}
              className={`flex flex-1 flex-col items-center rounded-xl py-2 ${
                rating === r.v ? 'bg-tg-button text-tg-button-text' : 'bg-tg-secondary text-tg-text'
              }`}
            >
              <span className="text-xl">{r.e}</span>
              <span className="text-[10px]">{r.l}</span>
            </button>
          ))}
        </div>

        <textarea
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="Комментарий (необязательно)"
          className="mt-4 w-full rounded-xl bg-tg-secondary p-3 text-tg-text outline-none"
          rows={3}
        />

        <div className="mt-4">
          <Button onClick={onSend} disabled={!lessonId || submit.isPending}>
            {submit.isPending ? 'Отправка...' : 'Отправить'}
          </Button>
        </div>
      </div>

      <div className="px-4 pb-24">
        <div className="mb-2 text-sm text-tg-hint">История</div>
        {history.data && history.data.length > 0 ? (
          history.data.map((h) => (
            <Card key={h.id} className="mb-2">
              <div className="flex items-center justify-between">
                <span className="font-medium">
                  {h.date} {h.time}
                </span>
                <span className="text-tg-hint">{h.note}</span>
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
