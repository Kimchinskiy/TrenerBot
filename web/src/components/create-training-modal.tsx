'use client'

import { useState, useMemo, useRef, useEffect } from 'react'
import { useClients, useGroups, useCreateScheduleEntry } from '@/lib/hooks'
import { Button } from '@/components/ui/button'
import { X, Search, Calendar, Clock, Users, Timer, Check, UserRound } from 'lucide-react'

interface Props {
  open: boolean
  onClose: () => void
}

export default function CreateTrainingModal({ open, onClose }: Props) {
  const { data: clients } = useClients()
  const { data: groups } = useGroups()
  const createEntry = useCreateScheduleEntry()
  const searchRef = useRef<HTMLInputElement>(null)

  const today = new Date().toISOString().slice(0, 10)
  const now = new Date()
  const defaultTime = `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes() + 60).padStart(2, '0')}`

  const [mode, setMode] = useState<'client' | 'group'>('client')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedClient, setSelectedClient] = useState<{ id: number; name: string } | null>(null)
  const [selectedGroup, setSelectedGroup] = useState<{ id: number; name: string } | null>(null)
  const [date, setDate] = useState(today)
  const [time, setTime] = useState(defaultTime)
  const [duration, setDuration] = useState('60')
  const [showResults, setShowResults] = useState(false)

  useEffect(() => {
    if (open) {
      setSearchQuery('')
      setSelectedClient(null)
      setSelectedGroup(null)
      setDate(today)
      setTime(defaultTime)
      setDuration('60')
      setMode('client')
    }
  }, [open])

  useEffect(() => {
    if (open && searchRef.current) {
      searchRef.current.focus()
    }
  }, [open])

  const filteredClients = useMemo(() => {
    if (!clients) return []
    if (!searchQuery.trim()) return clients
    const q = searchQuery.toLowerCase().trim()
    const seen = new Set<number>()
    return clients.filter((c) => {
      if (seen.has(c.id)) return false
      seen.add(c.id)
      return c.full_name.toLowerCase().includes(q)
    })
  }, [clients, searchQuery])

  const filteredGroups = useMemo(() => {
    if (!groups) return []
    if (!searchQuery.trim()) return groups
    const q = searchQuery.toLowerCase().trim()
    return groups.filter((g) => g.name && g.name.toLowerCase().includes(q) && g.active)
  }, [groups, searchQuery])

  const handleSelectClient = (id: number, name: string) => {
    setSelectedClient({ id, name })
    setSelectedGroup(null)
    setSearchQuery('')
    setShowResults(false)
  }

  const handleSelectGroup = (id: number, name: string) => {
    setSelectedGroup({ id, name })
    setSelectedClient(null)
    setSearchQuery('')
    setShowResults(false)
  }

  const handleSubmit = async () => {
    if (mode === 'client' && (!selectedClient || !date || !time)) return
    if (mode === 'group' && (!selectedGroup || !date || !time)) return
    createEntry.mutate(
      {
        ...(mode === 'client' ? { client_id: selectedClient!.id } : { group_id: selectedGroup!.id }),
        date,
        time,
        duration: Number(duration) || 60,
      },
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
          <div className="flex rounded-xl border border-border bg-muted/40 p-1">
            <button
              type="button"
              onClick={() => { setMode('client'); setSelectedGroup(null); setShowResults(false) }}
              className={`flex-1 flex items-center justify-center gap-2 rounded-lg py-2 text-sm font-bold transition-colors ${
                mode === 'client' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground'
              }`}
            >
              <UserRound className="h-4 w-4" /> Клиент
            </button>
            <button
              type="button"
              onClick={() => { setMode('group'); setSelectedClient(null); setShowResults(false) }}
              className={`flex-1 flex items-center justify-center gap-2 rounded-lg py-2 text-sm font-bold transition-colors ${
                mode === 'group' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground'
              }`}
            >
              <Users className="h-4 w-4" /> Группа
            </button>
          </div>

          <div className="relative">
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 flex items-center gap-1.5">
              {mode === 'client' ? <Users className="h-3.5 w-3.5" /> : <Users className="h-3.5 w-3.5" />}
              {mode === 'client' ? 'Клиент' : 'Группа'}
            </label>
            {mode === 'client' && selectedClient ? (
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
            ) : mode === 'group' && selectedGroup ? (
              <div className="flex items-center justify-between rounded-xl border border-border bg-background px-4 py-3">
                <div className="flex items-center gap-2">
                  <Check className="h-4 w-4 text-emerald-400" />
                  <span className="font-medium">{selectedGroup.name}</span>
                </div>
                <button
                  onClick={() => setSelectedGroup(null)}
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
                  placeholder={mode === 'client' ? 'Поиск по фамилии или имени...' : 'Поиск по названию группы...'}
                  className="w-full rounded-xl border border-border bg-background pl-10 pr-4 py-3 text-base"
                />
                {showResults && searchQuery.trim() && (
                  <div className="absolute z-10 mt-1 w-full rounded-xl border border-border bg-background shadow-xl max-h-48 overflow-y-auto">
                    {(mode === 'client' ? filteredClients : filteredGroups).length === 0 ? (
                      <div className="px-4 py-3 text-sm text-muted-foreground">Ничего не найдено</div>
                    ) : (
                      (mode === 'client' ? filteredClients : filteredGroups).map((item: any) => (
                        <button
                          key={item.id}
                          onClick={() => mode === 'client' ? handleSelectClient(item.id, item.full_name || item.name) : handleSelectGroup(item.id, item.name)}
                          className="w-full text-left px-4 py-3 text-sm hover:bg-muted/60 transition-colors border-b border-border/50 last:border-0"
                        >
                          {mode === 'client' ? (item as any).full_name : item.name}
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
            disabled={
              (mode === 'client' && (!selectedClient || !date || !time)) ||
              (mode === 'group' && (!selectedGroup || !date || !time)) ||
              createEntry.isPending
            }
            className="mt-2"
          >
            {createEntry.isPending ? 'Сохранение...' : 'Создать тренировку'}
          </Button>
        </div>
      </div>
    </div>
  )
}
