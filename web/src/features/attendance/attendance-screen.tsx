'use client'

import { useMemo, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useDateAttendance, useSaveDateAttendance } from '@/lib/hooks'
import { isoDate } from '@/lib/dates'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { haptics } from '@/services/telegram'
import type { DateAttendanceClient } from '@/lib/types'
import { CheckCircle2, Calendar } from 'lucide-react'

const WEEKDAYS_RU = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота']
const MONTHS_RU = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня', 'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря']

function fmtDate(iso: string): string {
  const [y, m, d] = iso.split('-').map(Number)
  const dt = new Date(y, m - 1, d)
  return `${d} ${MONTHS_RU[m - 1]}`
}

function fmtWeekday(iso: string): string {
  const [y, m, d] = iso.split('-').map(Number)
  const dt = new Date(y, m - 1, d)
  return WEEKDAYS_RU[dt.getDay()]
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

  const handleToggle = (clientId: number, checked: boolean) => {
    setDraft((prev) => ({ ...prev, [clientId]: checked }))
    setSuccessMsg('')
    setSaveError('')
  }

  const handleSave = async () => {
    setConfirming(false)
    setSaveError('')
    const entries = Object.entries(draft)
      .filter(([id]) => draft[+id] !== saved[+id])
      .map(([id, present]) => ({ client_id: +id, present }))

    if (entries.length === 0) return

    try {
      await saveMutation.mutateAsync({ date: today, entries })
      setSaved({ ...draft })
      setSuccessMsg('Посещаемость сохранена')
      setSaveError('')
      haptics.success()
    } catch {
      setSaveError('Ошибка сохранения')
      haptics.error()
    }
  }

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

  if (isLoading) {
    return (
      <div>
        <ScreenHeader title="Посещаемость" onBack={() => router.back()} />
        <Spinner label="Загрузка..." />
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <ScreenHeader title="Посещаемость" onBack={() => router.back()} />
        <ErrorBox error={error} />
      </div>
    )
  }

  return (
    <div>
      <ScreenHeader title="Посещаемость" onBack={() => router.back()} />

      <div className="px-5 pb-48">
        {/* Date */}
        <div className="flex items-center justify-center gap-2 mb-6 py-3 rounded-2xl bg-primary/5">
          <Calendar className="h-4 w-4 text-primary" />
          <span className="text-sm font-bold text-foreground">{fmtWeekday(today)}, {fmtDate(today)}</span>
        </div>

        {(!data || data.clients.length === 0) && (
          <Empty text="Сегодня нет тренировок" />
        )}

        {data && Array.from(clientsByTime.entries()).map(([time, clients]) => (
          <div key={time} className="mb-5">
            <div className="mb-2.5 text-xs font-bold uppercase tracking-wider text-muted-foreground px-1">{time}</div>
            <div className="flex flex-col gap-2">
              {clients.map((c) => (
                <div
                  key={c.client_id}
                  className="flex items-center gap-3.5 rounded-2xl bg-white p-4 shadow-card border border-border/30"
                >
                  <Avatar className="h-10 w-10 border border-border/50 shadow-sm shrink-0">
                    <AvatarFallback className="bg-primary/10 text-primary font-bold text-xs uppercase">
                      {c.full_name.charAt(0)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-bold text-foreground truncate">{c.full_name}</div>
                  </div>
                  <Switch
                    checked={draft[c.client_id] ?? false}
                    onCheckedChange={(checked) => handleToggle(c.client_id, checked)}
                  />
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Bottom action */}
      <div className="fixed bottom-[96px] left-0 right-0 z-40 border-t border-border/50 bg-white/95 backdrop-blur-md px-5 pt-4 pb-5">
        <div className="w-full max-w-md">
          {successMsg && (
            <div className="mb-3 flex items-center justify-center gap-2 text-sm font-semibold text-success">
              <CheckCircle2 className="h-4 w-4" />
              {successMsg}
            </div>
          )}
          {saveError && (
            <div className="mb-3 text-center text-sm font-semibold text-destructive">{saveError}</div>
          )}

          {confirming ? (
            <Card className="p-5 text-center shadow-elevated">
              <p className="text-base font-bold text-foreground mb-3">Сохранить посещаемость?</p>
              <div className="flex gap-3">
                <Button variant="outline" size="sm" onClick={() => setConfirming(false)} className="flex-1">
                  Отмена
                </Button>
                <Button size="sm" onClick={handleSave} disabled={saveMutation.isPending} className="flex-1">
                  {saveMutation.isPending ? 'Сохранение...' : 'Сохранить'}
                </Button>
              </div>
            </Card>
          ) : (
            <Button
              onClick={() => setConfirming(true)}
              disabled={!hasChanges || saveMutation.isPending}
              variant="gradient"
              size="lg"
            >
              {saveMutation.isPending ? 'Сохранение...' : `Сохранить посещаемость${dirtyCount > 0 ? ` (${dirtyCount})` : ''}`}
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
