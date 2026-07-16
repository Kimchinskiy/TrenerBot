'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAttendance, useClients, useMarkAttendance, useSchedule } from '@/lib/hooks'
import { weekRange, prettyDate } from '@/lib/dates'
import { Card, ScreenHeader, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { showMainButton, hideMainButton, haptics } from '@/services/telegram'

export default function AttendanceScreen() {
  const router = useRouter()
  const { from, to } = useMemo(() => weekRange(), [])
  const { data: lessons, isLoading, error } = useSchedule(from, to)
  const { data: clients } = useClients()
  const [selected, setSelected] = useState<number | null>(null)

  const { data: att, isLoading: attLoading } = useAttendance(selected ?? 0)
  const mark = useMarkAttendance(selected ?? 0)

  const nameOf = useMemo(() => {
    const m = new Map<number, string>()
    ;(clients || []).forEach((c) => m.set(c.id, c.full_name))
    return m
  }, [clients])

  const [states, setStates] = useState<Record<number, boolean>>({})
  const [original, setOriginal] = useState<Record<number, boolean>>({})

  useEffect(() => {
    if (!att) return
    const s: Record<number, boolean> = {}
    att.forEach((a) => (s[a.client_id] = a.present))
    setStates(s)
    setOriginal(s)
  }, [att])

  const dirty = useMemo(
    () => JSON.stringify(states) !== JSON.stringify(original),
    [states, original],
  )

  useEffect(() => {
    if (selected == null) {
      hideMainButton()
      return
    }
    showMainButton('Сохранить', async () => {
      const diffs = Object.entries(states).filter(([id]) => states[+id] !== original[+id])
      try {
        await Promise.all(
          diffs.map(([id, present]) => mark.mutateAsync({ clientId: +id, present })),
        )
        haptics.success()
      } catch {
        haptics.error()
      }
    })
    return () => hideMainButton()
  }, [selected, states, original, mark])

  if (!selected) {
    return (
      <div>
        <ScreenHeader title="Посещаемость" subtitle="Выберите тренировку" onBack={() => router.back()} />
        {isLoading && <Spinner label="Загрузка..." />}
        {error && <ErrorBox error={error} />}
        <div className="flex flex-col gap-2 px-4 pb-24">
          {(lessons || []).map((l) => (
            <Card key={l.id} className="flex items-center justify-between" onClick={() => setSelected(l.id)}>
              <div>
                <div className="font-semibold">
                  {prettyDate(l.date)} · {l.time}
                </div>
                {l.location && <div className="text-sm text-tg-hint">📍 {l.location}</div>}
              </div>
              <div className="text-tg-link">→</div>
            </Card>
          ))}
          {(lessons || []).length === 0 && <Empty text="Нет тренировок" />}
        </div>
      </div>
    )
  }

  return (
    <div>
      <ScreenHeader title="Посещаемость" subtitle="Отметьте ✓ / ✗" onBack={() => setSelected(null)} />
      {attLoading && <Spinner label="Загрузка..." />}
      <div className="flex flex-col gap-2 px-4 pb-24">
        {Object.keys(states).length === 0 && <Empty text="Нет записанных участников" />}
        {Object.entries(states).map(([cid, present]) => (
          <Card key={cid} className="flex items-center justify-between">
            <div className="font-medium">{nameOf.get(+cid) || `Участник #${cid}`}</div>
            <div className="flex gap-2">
              <button
                onClick={() => setStates((s) => ({ ...s, [cid]: true }))}
                className={`rounded-lg px-3 py-1 text-sm ${
                  present ? 'bg-green-600 text-white' : 'bg-tg-secondary text-tg-hint'
                }`}
              >
                ✓
              </button>
              <button
                onClick={() => setStates((s) => ({ ...s, [cid]: false }))}
                className={`rounded-lg px-3 py-1 text-sm ${
                  !present ? 'bg-red-600 text-white' : 'bg-tg-secondary text-tg-hint'
                }`}
              >
                ✗
              </button>
            </div>
          </Card>
        ))}
      </div>
    </div>
  )
}
