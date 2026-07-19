'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useDateAttendance, useSaveDateAttendance } from '@/lib/hooks'
import { isoDate, prettyDate } from '@/lib/dates'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { Button, Switch, Avatar, AvatarImage, AvatarFallback } from '@/components/ui'
import { haptics } from '@/services/telegram'
import type { DateAttendanceClient } from '@/lib/types'

const WEEKDAYS_RU = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота']
const MONTHS_RU = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня', 'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря']

function fmtDate(iso: string): string {
  const [y, m, d] = iso.split('-').map(Number)
  const dt = new Date(y, m - 1, d)
  return `${d} ${MONTHS_RU[m - 1]}\n${WEEKDAYS_RU[dt.getDay()]}`
}

export default function AttendanceScreen() {
  const router = useRouter()
  const today = useMemo(() => isoDate(new Date()), [])

  const { data, isLoading, error } = useDateAttendance(today)
  const saveMutation = useSaveDateAttendance()

  const [draft, setDraft] = useState<Record<number, boolean>>({})
  const [saved, setSaved] = useState<Record<number, boolean>>({})
  const [confirming, setConfirming] = useState(false)
  const [successMsg, setSuccessMsg] = useState('')
  const [saveError, setSaveError] = useState('')

  useEffect(() => {
    if (!data) return
    const s: Record<number, boolean> = {}
    data.clients.forEach((c) => {
      s[c.client_id] = c.present ?? false
    })
    setDraft(s)
    setSaved({ ...s })
  }, [data])

  const hasChanges = useMemo(
    () => JSON.stringify(draft) !== JSON.stringify(saved),
    [draft, saved],
  )

  const dirtyCount = useMemo(() => {
    let n = 0
    for (const id of Object.keys(draft)) {
      if (draft[+id] !== saved[+id]) n++
    }
    return n
  }, [draft, saved])

  const handleToggle = useCallback((clientId: number, checked: boolean) => {
    setDraft((prev) => ({ ...prev, [clientId]: checked }))
    setSuccessMsg('')
    setSaveError('')
  }, [])

  const handleSave = useCallback(async () => {
    setConfirming(false)
    setSaveError('')
    const entries = Object.entries(draft)
      .filter(([id]) => draft[+id] !== saved[+id])
      .map(([id, present]) => ({ client_id: +id, present }))

    if (entries.length === 0) return

    try {
      await saveMutation.mutateAsync({ date: today, entries })
      setSaved({ ...draft })
      setSuccessMsg('Посещаемость успешно сохранена.')
      setSaveError('')
      haptics.success()
    } catch {
      setSaveError('Ошибка сохранения. Попробуйте снова.')
      haptics.error()
    }
  }, [draft, saved, today, saveMutation])

  // Group clients by time
  const clientsByTime = useMemo(() => {
    const map = new Map<string, DateAttendanceClient[]>()
    if (!data) return map
    data.clients.forEach((c) => {
      const arr = map.get(c.time) || []
      arr.push(c)
      map.set(c.time, arr)
    })
    return map
  }, [data])

  if (isLoading) return <div><ScreenHeader title="Посещаемость" onBack={() => router.back()} /><Spinner label="Загрузка..." /></div>
  if (error) return <div><ScreenHeader title="Посещаемость" onBack={() => router.back()} /><ErrorBox error={error} /></div>

  return (
    <div>
      <ScreenHeader title="Посещаемость" onBack={() => router.back()} />

      <div className="px-4 pb-48">
        {/* Date header */}
        <div className="mb-6 text-center">
          <div className="text-xl font-bold tracking-tight text-foreground">{fmtDate(today)}</div>
        </div>

        {/* Clients */}
        {(!data || data.clients.length === 0) && <Empty text="Сегодня нет тренировок" />}

        {data && Array.from(clientsByTime.entries()).map(([time, clients]) => (
          <div key={time} className="mb-6">
            <div className="mb-3 text-xs font-bold uppercase tracking-wider text-muted-foreground px-1">{time}</div>
            <div className="flex flex-col gap-3">
              {clients.map((c) => (
                <Card key={c.client_id} className="flex items-center gap-3.5 py-4 px-5 shadow-sm border-border/80 hover:border-foreground/10 transition-all duration-200">
                  <Avatar className="h-10 w-10 border border-border shadow-sm shrink-0">
                    {c.photo ? (
                      <AvatarImage src={c.photo} alt={c.full_name} />
                    ) : (
                      <AvatarFallback className="bg-primary/10 text-primary font-bold uppercase">{c.full_name.charAt(0)}</AvatarFallback>
                    )}
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <div className="text-base font-bold text-foreground truncate">{c.full_name}</div>
                  </div>
                  <Switch
                    checked={draft[c.client_id] ?? false}
                    onCheckedChange={(checked) => handleToggle(c.client_id, checked)}
                  />
                </Card>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Bottom action bar — pinned above the tab bar */}
      <div className="fixed bottom-[96px] left-0 right-0 z-40 border-t border-border bg-background/95 backdrop-blur-md px-4 pt-4 pb-5 flex flex-col items-center">
        <div className="w-full max-w-md">
          {successMsg && (
            <div className="mb-3 text-center text-sm font-semibold text-emerald-400">{successMsg}</div>
          )}
          {saveError && (
            <div className="mb-3 text-center text-sm font-semibold text-destructive">{saveError}</div>
          )}

          {confirming ? (
            <Card className="p-5 text-center shadow-lg border-border bg-card/90">
              <div className="text-base font-bold mb-3 text-foreground">Сохранить посещаемость за сегодня?</div>
              <div className="flex gap-3 justify-center">
                <Button variant="outline" size="sm" onClick={() => setConfirming(false)}>Отмена</Button>
                <Button size="sm" onClick={handleSave} disabled={saveMutation.isPending}>
                  {saveMutation.isPending ? 'Сохранение...' : 'Сохранить'}
                </Button>
              </div>
            </Card>
          ) : (
            <Button onClick={() => setConfirming(true)} disabled={!hasChanges || saveMutation.isPending} className="font-bold">
              {saveMutation.isPending ? 'Сохранение...' : `Сохранить посещаемость${dirtyCount > 0 ? ` (${dirtyCount})` : ''}`}
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
