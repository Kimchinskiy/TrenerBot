'use client'

import { useMemo, useState } from 'react'
import { useMe, useScheduleWeek } from '@/lib/hooks'
import { isoDate, getWeekDays, MONTHS_RU_NOM, CalendarDay, prettyDateLong } from '@/lib/dates'
import { Card } from '@/components/ui/card'
import { SkeletonList } from '@/components/ui/skeleton'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import type { ScheduleEntry } from '@/lib/types'
import { Plus, Clock, Users, ChevronLeft, ChevronRight, Calendar as CalendarIcon, MapPin } from 'lucide-react'
import CreateTrainingModal from '@/components/create-training-modal'

function LessonCard({ entry }: { entry: ScheduleEntry }) {
  const isCanceled = entry.status === 'canceled'

  // Calculate end time
  const [h, m] = entry.time.split(':').map(Number)
  const startTime = `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
  const endMinutes = h * 60 + m + (entry.duration || 60)
  const endH = Math.floor(endMinutes / 60) % 24
  const endM = endMinutes % 60
  const endTime = `${String(endH).padStart(2, '0')}:${String(endM).padStart(2, '0')}`

  return (
    <Card
      className={`flex flex-col gap-3 rounded-3xl p-4 shadow-card transition-all duration-200 ${
        isCanceled
          ? 'bg-destructive/5 border border-destructive/20 opacity-75'
          : 'bg-card border border-border/50 hover:shadow-elevated'
      }`}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
          <Clock className="h-3.5 w-3.5 text-primary" />
          <span>{startTime} – {endTime}</span>
        </div>
        <span
          className={`text-xs font-semibold px-3 py-1 rounded-xl shadow-sm ${
            isCanceled
              ? 'bg-destructive/10 text-destructive'
              : 'bg-primary/10 text-primary border border-primary/20'
          }`}
        >
          {isCanceled ? 'Отменено' : (entry.group_name ? 'Группа' : 'Зал А')}
        </span>
      </div>

      <div className="flex items-center gap-3">
        <Avatar className="h-11 w-11 border border-border/50 shadow-sm shrink-0">
          <AvatarFallback
            className={`text-sm font-bold uppercase ${
              isCanceled ? 'bg-destructive/10 text-destructive' : 'bg-primary/10 text-primary'
            }`}
          >
            {entry.group_name ? <Users className="h-5 w-5" /> : (entry.client_name?.charAt(0) || '?')}
          </AvatarFallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <h3 className="text-base font-bold text-foreground truncate">
            {entry.group_name || entry.client_name}
          </h3>
          <div className="flex items-center gap-1 text-xs text-muted-foreground mt-0.5">
            <MapPin className="h-3 w-3 text-muted-foreground/70" />
            <span>Бассейн 25 м</span>
            <span className="mx-1">•</span>
            <span>{entry.duration || 60} мин</span>
          </div>
        </div>
      </div>
    </Card>
  )
}

function CoachView() {
  const [currentDate, setCurrentDate] = useState(new Date())
  const [selectedDateIso, setSelectedDateIso] = useState(isoDate(new Date()))
  const [showCreate, setShowCreate] = useState(false)

  // Get 7 week days
  const weekDays = useMemo(() => getWeekDays(currentDate), [currentDate])

  const fromDate = weekDays[0].iso
  const toDate = weekDays[6].iso

  const { data: scheduleData, isLoading, error } = useScheduleWeek(fromDate, toDate)

  // Entries for selected date
  const selectedEntries = useMemo(() => {
    if (!scheduleData) return []
    return scheduleData
      .filter((e) => e.date === selectedDateIso)
      .sort((a, b) => a.time.localeCompare(b.time))
  }, [scheduleData, selectedDateIso])

  // Map of dates that have entries for indicator dots
  const entryDatesSet = useMemo(() => {
    const set = new Set<string>()
    if (scheduleData) {
      scheduleData.forEach((e) => set.add(e.date))
    }
    return set
  }, [scheduleData])

  const prevWeek = () => {
    const prev = new Date(currentDate)
    prev.setDate(prev.getDate() - 7)
    setCurrentDate(prev)
    setSelectedDateIso(isoDate(prev))
  }

  const nextWeek = () => {
    const next = new Date(currentDate)
    next.setDate(next.getDate() + 7)
    setCurrentDate(next)
    setSelectedDateIso(isoDate(next))
  }

  const currentMonthName = weekDays[2]?.monthName || MONTHS_RU_NOM[currentDate.getMonth()]

  return (
    <>
      <CreateTrainingModal open={showCreate} onClose={() => setShowCreate(false)} />

      <div className="px-5 pb-24 flex flex-col gap-4">
        {/* Top Header: Month & Navigation Arrows */}
        <div className="flex items-center justify-between pt-2 px-1">
          <h2 className="text-xl font-bold text-foreground">
            {currentMonthName}
          </h2>
          <div className="flex items-center gap-1.5">
            <button
              onClick={prevWeek}
              className="flex h-9 w-9 items-center justify-center rounded-xl bg-card border border-border/50 shadow-sm text-foreground hover:bg-muted/50 transition-colors"
            >
              <ChevronLeft className="h-5 w-5" />
            </button>
            <button
              onClick={nextWeek}
              className="flex h-9 w-9 items-center justify-center rounded-xl bg-card border border-border/50 shadow-sm text-foreground hover:bg-muted/50 transition-colors"
            >
              <ChevronRight className="h-5 w-5" />
            </button>
          </div>
        </div>

        {/* Horizontal Calendar Week Strip */}
        <Card className="p-3 bg-card border border-border/50 shadow-card rounded-3xl">
          <div className="grid grid-cols-7 text-center">
            {weekDays.map((day) => {
              const isSelected = day.iso === selectedDateIso
              const hasEntries = entryDatesSet.has(day.iso)
              const isWeekend = day.dayName === 'Сб' || day.dayName === 'Вс'

              return (
                <button
                  key={day.iso}
                  onClick={() => setSelectedDateIso(day.iso)}
                  className="flex flex-col items-center gap-1.5 py-1 transition-all"
                >
                  <span
                    className={`text-xs font-semibold ${
                      isSelected
                        ? 'text-primary font-bold'
                        : isWeekend
                        ? 'text-rose-500 font-bold'
                        : 'text-muted-foreground'
                    }`}
                  >
                    {day.dayName}
                  </span>
                  <div
                    className={`flex h-10 w-10 items-center justify-center rounded-2xl text-sm font-bold transition-all duration-200 ${
                      isSelected
                        ? 'bg-primary text-white shadow-md shadow-primary/25 scale-105'
                        : 'text-foreground hover:bg-muted/40'
                    }`}
                  >
                    {day.dayNumber}
                  </div>
                  {/* Indicator Dot */}
                  <div className="h-1 w-1 rounded-full">
                    {hasEntries && (
                      <span className={`block h-1 w-1 rounded-full ${isSelected ? 'bg-primary' : 'bg-primary/60'}`} />
                    )}
                  </div>
                </button>
              )
            })}
          </div>
        </Card>

        {/* Trainings List for Selected Day */}
        <div className="flex flex-col gap-3 mt-1">
          {isLoading && <SkeletonList count={2} />}

          {!isLoading && selectedEntries.length === 0 && (
            <div className="flex flex-col items-center justify-center py-10 px-4 text-center rounded-3xl border border-dashed border-border/60 bg-card/40">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary mb-3">
                <CalendarIcon className="h-6 w-6" />
              </div>
              <p className="text-sm font-semibold text-foreground mb-1">
                Нет тренировок на выбранный день
              </p>
              <p className="text-xs text-muted-foreground max-w-xs">
                Нажмите на кнопку ниже, чтобы запланировать новую тренировку
              </p>
            </div>
          )}

          {!isLoading &&
            selectedEntries.map((entry) => (
              <LessonCard key={entry.id} entry={entry} />
            ))}
        </div>
      </div>

      {/* FAB Create Training Button */}
      <button
        onClick={() => setShowCreate(true)}
        className="fixed bottom-24 right-5 z-40 h-14 w-14 rounded-2xl bg-gradient-to-r from-primary to-cyan-500 text-white shadow-glow flex items-center justify-center active:scale-90 transition-transform"
      >
        <Plus className="h-6 w-6" />
      </button>
    </>
  )
}

function ParentView() {
  const meDate = new Date()
  const [selectedDateIso, setSelectedDateIso] = useState(isoDate(meDate))
  const weekDays = useMemo(() => getWeekDays(meDate), [meDate])
  const fromDate = weekDays[0].iso
  const toDate = weekDays[6].iso

  const { data: scheduleData, isLoading } = useScheduleWeek(fromDate, toDate)

  const selectedEntries = useMemo(() => {
    if (!scheduleData) return []
    return scheduleData
      .filter((e) => e.date === selectedDateIso)
      .sort((a, b) => a.time.localeCompare(b.time))
  }, [scheduleData, selectedDateIso])

  return (
    <div className="px-5 pb-24 flex flex-col gap-4">
      <div className="flex items-center justify-between pt-2 px-1">
        <h2 className="text-xl font-bold text-foreground">
          {weekDays[2]?.monthName}
        </h2>
      </div>

      <Card className="p-3 bg-card border border-border/50 shadow-card rounded-3xl">
        <div className="grid grid-cols-7 text-center">
          {weekDays.map((day) => {
            const isSelected = day.iso === selectedDateIso
            const isWeekend = day.dayName === 'Сб' || day.dayName === 'Вс'
            return (
              <button
                key={day.iso}
                onClick={() => setSelectedDateIso(day.iso)}
                className="flex flex-col items-center gap-1.5 py-1 transition-all"
              >
                <span
                  className={`text-xs font-semibold ${
                    isSelected
                      ? 'text-primary font-bold'
                      : isWeekend
                      ? 'text-rose-500 font-bold'
                      : 'text-muted-foreground'
                  }`}
                >
                  {day.dayName}
                </span>
                <div
                  className={`flex h-10 w-10 items-center justify-center rounded-2xl text-sm font-bold transition-all ${
                    isSelected
                      ? 'bg-primary text-white shadow-md shadow-primary/25 scale-105'
                      : 'text-foreground hover:bg-muted/40'
                  }`}
                >
                  {day.dayNumber}
                </div>
              </button>
            )
          })}
        </div>
      </Card>

      <div className="flex flex-col gap-3">
        {isLoading && <SkeletonList count={2} />}
        {!isLoading && selectedEntries.length === 0 && (
          <div className="p-8 text-center text-sm text-muted-foreground border border-dashed border-border/60 rounded-3xl">
            На этот день нет занятий
          </div>
        )}
        {!isLoading &&
          selectedEntries.map((entry) => (
            <LessonCard key={entry.id} entry={entry} />
          ))}
      </div>
    </div>
  )
}

export default function Schedule() {
  const { data: me } = useMe()

  return (
    <div className="pt-2">
      {me?.role === 'parent' ? <ParentView /> : <CoachView />}
    </div>
  )
}
