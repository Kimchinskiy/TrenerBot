'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useGroups, useCreateGroup, useUpdateGroup, useDeleteGroup } from '@/lib/hooks'
import { ScreenHeader, Spinner, ErrorBox, Empty } from '@/components/ui/screen'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus, Edit3, Trash2, Users, X, Check, Calendar, ChevronRight } from 'lucide-react'
import type { Group } from '@/lib/types'

function GroupCard({
  group,
  onEdit,
  onDelete,
  onOpen,
}: {
  group: Group
  onEdit: (g: Group) => void
  onDelete: (g: Group) => void
  onOpen: (g: Group) => void
}) {
  return (
    <Card
      className="p-4 cursor-pointer transition-all hover:shadow-elevated border-border/50 bg-card active:scale-[0.99] flex flex-col gap-3"
      onClick={() => onOpen(group)}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-3 min-w-0">
          <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 text-primary shrink-0 shadow-sm border border-primary/15">
            <Users className="h-5.5 w-5.5" />
          </div>
          <div className="min-w-0">
            <h3 className="text-base font-bold text-foreground truncate">{group.name || 'Без названия'}</h3>
            {group.location && (
              <p className="text-xs text-muted-foreground truncate mt-0.5 flex items-center gap-1">
                <span>📍</span> {group.location}
              </p>
            )}
          </div>
        </div>

        <div className="flex items-center gap-1 shrink-0" onClick={(e) => e.stopPropagation()}>
          <button
            onClick={() => onEdit(group)}
            className="rounded-xl p-2 hover:bg-muted/70 text-muted-foreground hover:text-foreground transition-colors"
            title="Редактировать"
          >
            <Edit3 className="h-4 w-4" />
          </button>
          <button
            onClick={() => onDelete(group)}
            className="rounded-xl p-2 hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
            title="Удалить"
          >
            <Trash2 className="h-4 w-4 text-destructive" />
          </button>
          <div className="pl-1 text-muted-foreground/60">
            <ChevronRight className="h-4 w-4" />
          </div>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-border/40">
        {group.schedule && (
          <span className="inline-flex items-center gap-1.5 rounded-xl bg-muted/70 px-2.5 py-1 text-xs font-medium text-foreground">
            <Calendar className="h-3.5 w-3.5 text-primary" />
            {group.schedule}
          </span>
        )}
        <span className="inline-flex items-center gap-1.5 rounded-xl bg-primary/10 px-2.5 py-1 text-xs font-semibold text-primary">
          <Users className="h-3.5 w-3.5" />
          {group.member_count ?? 0}
          {group.max_members ? ` / ${group.max_members}` : ''} учеников
        </span>
        {group.price && (
          <span className="inline-flex items-center rounded-xl bg-emerald-500/10 px-2.5 py-1 text-xs font-bold text-emerald-700 dark:text-emerald-400 ml-auto">
            {group.price.toLocaleString('ru-RU')} ₽
          </span>
        )}
      </div>
    </Card>
  )
}

function CreateForm({
  onSubmit,
  initial,
  onCancel,
  isPending,
}: {
  onSubmit: (data: { name: string; schedule?: string; max_members?: number; price?: number; location?: string }) => void
  initial?: Partial<Group>
  onCancel: () => void
  isPending: boolean
}) {
  const [name, setName] = useState(initial?.name || '')
  const [schedule, setSchedule] = useState(initial?.schedule || '')
  const [maxMembers, setMaxMembers] = useState(initial?.max_members?.toString() || '')
  const [price, setPrice] = useState(initial?.price?.toString() || '')

  return (
    <Card className="p-5">
      <div className="flex flex-col gap-3">
        <div>
          <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block">
            Название группы
          </label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Название"
            className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-2.5 text-sm"
          />
        </div>
        <div>
          <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block">
            Расписание
          </label>
          <input
            value={schedule}
            onChange={(e) => setSchedule(e.target.value)}
            placeholder="Например: ПН/СР/ПТ 10:00"
            className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-2.5 text-sm"
          />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block">
              Кол-во мест
            </label>
            <input
              type="number"
              value={maxMembers}
              onChange={(e) => setMaxMembers(e.target.value)}
              placeholder="10"
              className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-2.5 text-sm"
            />
          </div>
          <div>
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block">
              Цена (₽)
            </label>
            <input
              type="number"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              placeholder="5000"
              className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-2.5 text-sm"
            />
          </div>
        </div>
        <div className="flex gap-2 mt-1">
          <Button variant="outline" size="sm" onClick={onCancel} className="flex-1">
            Отмена
          </Button>
          <Button
            size="sm"
            onClick={() =>
              onSubmit({
                name: name.trim(),
                schedule: schedule.trim() || undefined,
                max_members: maxMembers ? Number(maxMembers) : undefined,
                price: price ? Number(price) : undefined,
              })
            }
            disabled={!name.trim() || isPending}
            variant="gradient"
            className="flex-1"
          >
            {isPending ? 'Сохранение...' : initial ? 'Сохранить' : 'Создать'}
          </Button>
        </div>
      </div>
    </Card>
  )
}

export default function GroupsScreen() {
  const router = useRouter()
  const { data: groups, isLoading, error } = useGroups()
  const createGroup = useCreateGroup()
  const updateGroup = useUpdateGroup()
  const deleteGroup = useDeleteGroup()

  const [showCreate, setShowCreate] = useState(false)
  const [editing, setEditing] = useState<Group | null>(null)
  const [deleting, setDeleting] = useState<Group | null>(null)
  const [doneMsg, setDoneMsg] = useState('')

  const handleCreate = (data: { name: string; schedule?: string; max_members?: number; price?: number }) => {
    createGroup.mutate(data, {
      onSuccess: () => {
        setShowCreate(false)
        setDoneMsg('Группа создана')
        setTimeout(() => setDoneMsg(''), 3000)
      },
    })
  }

  const handleUpdate = (data: { name: string; schedule?: string; max_members?: number; price?: number }) => {
    if (!editing) return
    updateGroup.mutate({ id: editing.id, ...data }, {
      onSuccess: () => {
        setEditing(null)
        setDoneMsg('Группа обновлена')
        setTimeout(() => setDoneMsg(''), 3000)
      },
    })
  }

  const handleDelete = () => {
    if (!deleting) return
    deleteGroup.mutate(deleting.id, {
      onSuccess: () => {
        setDeleting(null)
        setDoneMsg('Группа удалена')
        setTimeout(() => setDoneMsg(''), 3000)
      },
    })
  }

  return (
    <div className="pt-4">
      <div className="px-5 pb-24 flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
            {groups?.length ? `${groups.length} групп` : 'Список групп'}
          </span>
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 rounded-xl bg-primary px-3 py-1.5 text-xs font-semibold text-white shadow-sm hover:opacity-90 active:scale-95 transition-all"
          >
            <Plus className="h-4 w-4" />
            <span>Новая группа</span>
          </button>
        </div>

        {doneMsg && (
          <div className="flex items-center justify-center gap-2 rounded-2xl bg-success-light/50 px-4 py-3 text-sm font-semibold text-success">
            <Check className="h-4 w-4" />
            {doneMsg}
          </div>
        )}

        {showCreate && (
          <CreateForm
            onSubmit={handleCreate}
            onCancel={() => setShowCreate(false)}
            isPending={createGroup.isPending}
          />
        )}

        {editing && (
          <CreateForm
            onSubmit={handleUpdate}
            initial={editing}
            onCancel={() => setEditing(null)}
            isPending={updateGroup.isPending}
          />
        )}

        {deleting && (
          <Card className="p-5 text-center">
            <p className="text-base font-bold text-foreground mb-1">
              Удалить группу «{deleting.name}»?
            </p>
            <p className="text-sm text-muted-foreground mb-4">
              Это действие нельзя отменить
            </p>
            <div className="flex gap-3">
              <Button variant="outline" size="sm" onClick={() => setDeleting(null)} className="flex-1">
                Отмена
              </Button>
              <Button
                size="sm"
                variant="destructive"
                onClick={handleDelete}
                disabled={deleteGroup.isPending}
                className="flex-1"
              >
                {deleteGroup.isPending ? 'Удаление...' : 'Удалить'}
              </Button>
            </div>
          </Card>
        )}

        {isLoading && <Spinner label="Загрузка групп..." />}
        {error && <ErrorBox error={error} />}
        {!isLoading && !error && (!groups || groups.length === 0) && !showCreate && (
          <Empty
            text="У вас ещё нет групп"
            icon={<Users className="h-8 w-8" />}
          />
        )}

        {groups && groups.length > 0 && (
          <div className="flex flex-col gap-2">
            {groups.map((g) => (
              <GroupCard
                key={g.id}
                group={g}
                onEdit={setEditing}
                onDelete={setDeleting}
                onOpen={(g) => router.push(`/dashboard/groups/${g.id}`)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
