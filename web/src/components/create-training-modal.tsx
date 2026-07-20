'use client'

import { useState, useMemo, useRef, useEffect } from 'react'
import { useClients, useCreateScheduleEntry } from '@/lib/hooks'
import { Button } from '@/components/ui/button'
import { X, Search, Calendar, Clock, Users, Timer, Check } from 'lucide-react'

interface Props {
  open: boolean
  onClose: () => void
}

export default function CreateTrainingModal({ open, onClose }: Props) {
  const { data: clients } = useClients()
  const createEntry = useCreateScheduleEntry()
  const searchRef = useRef<HTMLInputElement>(null)

  const today = new Date().toISOString().slice(0, 10)
  const now = new Date()
  const defaultTime = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes() + 60).padStart(2, '0')}`

  const [searchQuery, setSearchQuery] = useState('')
  const [selectedClient, setSelectedClient] = useState<{ id: number; name: string } | null>(null)
  const [date, setDate] = useState(today)
  const [time, setTime] = useState(defaultTime)
  const [duration, setDuration] = useState('60')
  const [showResults, setShowResults] = useState(false)

  useEffect(() => {
    if (open) {
      setSearchQuery('')
      setSelectedClient(null)
      setDate(today)
      setTime(defaultTime)
      setDuration('60')
    }
  }, [open])

  useEffect(() => {
    if (open && searchRef.current) {
      searchRef.current.focus()
    }
  }, [open])

  const filtered = useMemo(() => {
    if (!clients) return []
    if (!searchQuery.trim()) return clients
    const q = searchQuery.toLowerCase().trim()
    return clients.filter((c) => c.full_name.toLowerCase().includes(q))
  }, [clients, searchQuery])

  const handleSelect = (id: number, name: string) => {
    setSelectedClient({ id, name })
    setSearchQuery('')
    setShowResults(false)
  }

  const handleSubmit = async () => {
    if (!selectedClient || !date || !time) return
    createEntry.mutate(
      { client_id: selectedClient.id, date, time, duration: Number(duration) || 60 },
      { onSuccess: () => onClose() },
    )
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" onClick={onClose}>
      <div
        className="w-full max-w-lg rounded-2xl bg-background p-5 pb-8 shadow-2xl animate-in fade-in zoom-in-95"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-bold">Новая тренировка</h2>
          <button onClick={onClose} className="rounded-full p-2 hover:bg-muted/60 transition-colors">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="flex flex-col gap-4">
          <div className="relative">
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 flex items-center gap-1.5">
              <Users className="h-3.5 w-3.5" /> Клиент
            </label>
            {selectedClient ? (
              <div className="flex items-center justify-between rounded-xl border border-border bg-background px-4 py-3">
                <div className="flex items-center gap-2">
                  <Check className="h-4 w-4 text-emerald-400" />
                  <span className="font-medium">{selectedClient.name}</span>
                </div>
                <button
                  onClick={() => setSelectedClient(null)}
                  className="text-xs text-muted-foreground hover:text-foreground underline"
                >
                  Изменить
                </button>
              </div>
            ) : (
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <input
                  ref={searchRef}
                  type="text"
                  value={searchQuery}
                  onChange={(e) => { setSearchQuery(e.target.value); setShowResults(true) }}
                  onFocus={() => setShowResults(true)}
                  placeholder="Поиск по фамилии или имени..."
                  className="w-full rounded-xl border border-border bg-background pl-10 pr-4 py-3 text-base"
                />
                {showResults && searchQuery.trim() && (
                  <div className="absolute z-10 mt-1 w-full rounded-xl border border-border bg-background shadow-xl max-h-48 overflow-y-auto">
                    {filtered.length === 0 ? (
                      <div className="px-4 py-3 text-sm text-muted-foreground">Ничего не найдено</div>
                    ) : (
                      filtered.map((c) => (
                        <button
                          key={c.id}
                          onClick={() => handleSelect(c.id, c.full_name)}
                          className="w-full text-left px-4 py-3 text-sm hover:bg-muted/60 transition-colors border-b border-border/50 last:border-0"
                        >
                          {c.full_name}
                        </button>
                      ))
                    )}
                  </div>
                )}
              </div>
            )}
          </div>

          <div>
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 flex items-center gap-1.5">
              <Calendar className="h-3.5 w-3.5" /> Дата
            </label>
            <input
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              className="w-full rounded-xl border border-border bg-background px-4 py-3 text-base"
            />
          </div>

          <div>
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 flex items-center gap-1.5">
              <Clock className="h-3.5 w-3.5" /> Время
            </label>
            <input
              type="time"
              value={time}
              onChange={(e) => setTime(e.target.value)}
              className="w-full rounded-xl border border-border bg-background px-4 py-3 text-base"
            />
          </div>

          <div>
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 flex items-center gap-1.5">
              <Timer className="h-3.5 w-3.5" /> Длительность (мин)
            </label>
            <input
              type="number"
              min={15}
              max={180}
              step={5}
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
              className="w-full rounded-xl border border-border bg-background px-4 py-3 text-base"
            />
          </div>

          <Button
            onClick={handleSubmit}
            disabled={!selectedClient || !date || !time || createEntry.isPending}
            className="mt-2"
          >
            {createEntry.isPending ? 'Сохранение...' : 'Создать тренировку'}
          </Button>
        </div>
      </div>
    </div>
  )
}
