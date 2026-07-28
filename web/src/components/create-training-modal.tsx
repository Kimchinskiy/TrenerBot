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
    return clients.filter((c) => c.full_name.toLowerCase().includes(q))
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
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" onClick={onClose}>
      <div
        className="w-full max-w-lg rounded-3xl bg-card border border-border/60 p-5 pb-8 shadow-elevated fade-in"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-bold text-foreground">Новая тренировка</h2>
          <button onClick={onClose} className="rounded-xl p-2 hover:bg-muted/50 transition-colors">
            <X className="h-5 w-5 text-muted-foreground" />
          </button>
        </div>

        <div className="flex flex-col gap-4">
          <div className="flex rounded-2xl bg-muted/50 p-1">
            <button
              type="button"
              onClick={() => { setMode('client'); setSelectedGroup(null); setShowResults(false) }}
              className={`flex-1 flex items-center justify-center gap-2 rounded-xl py-2.5 text-sm font-semibold transition-all duration-200 ${
                mode === 'client' ? 'bg-card shadow-sm text-foreground' : 'text-muted-foreground'
              }`}
            >
              <UserRound className="h-4 w-4" /> Клиент
            </button>
            <button
              type="button"
              onClick={() => { setMode('group'); setSelectedClient(null); setShowResults(false) }}
              className={`flex-1 flex items-center justify-center gap-2 rounded-xl py-2.5 text-sm font-semibold transition-all duration-200 ${
                mode === 'group' ? 'bg-card shadow-sm text-foreground' : 'text-muted-foreground'
              }`}
            >
              <Users className="h-4 w-4" /> Группа
            </button>
          </div>

          <div className="relative">
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block">
              {mode === 'client' ? 'Клиент' : 'Группа'}
            </label>
            {mode === 'client' && selectedClient ? (
              <div className="flex items-center justify-between rounded-2xl border border-border/60 bg-card px-4 py-3">
                <div className="flex items-center gap-2">
                  <Check className="h-4 w-4 text-success" />
                  <span className="text-sm font-semibold text-foreground">{selectedClient.name}</span>
                </div>
                <button onClick={() => setSelectedClient(null)} className="text-xs text-primary font-semibold">
                  Изменить
                </button>
              </div>
            ) : mode === 'group' && selectedGroup ? (
              <div className="flex items-center justify-between rounded-2xl border border-border/60 bg-card px-4 py-3">
                <div className="flex items-center gap-2">
                  <Check className="h-4 w-4 text-success" />
                  <span className="text-sm font-semibold text-foreground">{selectedGroup.name}</span>
                </div>
                <button onClick={() => setSelectedGroup(null)} className="text-xs text-primary font-semibold">
                  Изменить
                </button>
              </div>
            ) : (
              <div className="relative">
                <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/60" />
                <input
                  ref={searchRef}
                  type="text"
                  value={searchQuery}
                  onChange={(e) => { setSearchQuery(e.target.value); setShowResults(true) }}
                  onFocus={() => setShowResults(true)}
                  placeholder={mode === 'client' ? 'Поиск по фамилии...' : 'Поиск по названию...'}
                  className="w-full rounded-2xl border border-border/60 bg-card text-foreground placeholder:text-muted-foreground/60 pl-10 pr-4 py-3 text-sm shadow-sm"
                />
                {mode === 'group' && !selectedGroup && !searchQuery.trim() && groups && groups.length === 0 && (
                  <div className="mt-2 rounded-2xl border border-dashed border-border bg-muted/30 px-4 py-3 text-center">
                    <p className="text-sm font-medium text-muted-foreground">У вас ещё нет групп</p>
                    <p className="text-xs text-muted-foreground mt-1">Создайте группу в разделе «Группы»</p>
                  </div>
                )}
                {showResults && searchQuery.trim() && (
                  <div className="absolute z-10 mt-1 w-full rounded-2xl border border-border/50 bg-card text-foreground shadow-elevated max-h-48 overflow-y-auto">
                    {(mode === 'client' ? filteredClients : filteredGroups).length === 0 ? (
                      <div className="px-4 py-3 text-sm text-muted-foreground">Ничего не найдено</div>
                    ) : (
                      (mode === 'client' ? filteredClients : filteredGroups).map((item: any) => (
                        <button
                          key={item.id}
                          onClick={() => mode === 'client' ? handleSelectClient(item.id, item.full_name) : handleSelectGroup(item.id, item.name)}
                          className="w-full text-left px-4 py-3 text-sm hover:bg-muted/40 transition-colors border-b border-border/30 last:border-0 font-medium text-foreground"
                        >
                          {mode === 'client' ? item.full_name : item.name}
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
              className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-3 text-sm shadow-sm"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 flex items-center gap-1.5">
                <Clock className="h-3.5 w-3.5" /> Время
              </label>
              <input
                type="time"
                value={time}
                onChange={(e) => setTime(e.target.value)}
                className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-3 text-sm shadow-sm"
              />
            </div>
            <div>
              <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 flex items-center gap-1.5">
                <Timer className="h-3.5 w-3.5" /> Длительность
              </label>
              <input
                type="number"
                min={15}
                max={180}
                step={5}
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
                className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-3 text-sm shadow-sm"
              />
            </div>
          </div>

          <Button
            onClick={handleSubmit}
            variant="gradient"
            size="lg"
            disabled={
              (mode === 'client' && (!selectedClient || !date || !time)) ||
              (mode === 'group' && (!selectedGroup || !date || !time)) ||
              createEntry.isPending
            }
          >
            {createEntry.isPending ? 'Сохранение...' : 'Создать тренировку'}
          </Button>
        </div>
      </div>
    </div>
  )
}
