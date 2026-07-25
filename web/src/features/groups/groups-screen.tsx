'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useGroups, useCreateGroup, useUpdateGroup, useDeleteGroup } from '@/lib/hooks'
import { ScreenHeader, Spinner, ErrorBox, Empty } from '@/components/ui/screen'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus, Edit3, Trash2, Users, X, Check } from 'lucide-react'
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
    <Card className="flex items-center gap-4 cursor-pointer transition-colors hover:bg-muted/20" onClick={() => onOpen(group)}>
      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 shrink-0">
        <Users className="h-5 w-5 text-primary" />
      </div>
      <div className="flex-1 min-w-0">
        <h3 className="text-sm font-bold text-foreground truncate">{group.name || 'Без названия'}</h3>
        <div className="flex items-center gap-2 mt-0.5">
          {group.schedule && (
            <span className="text-xs text-muted-foreground">{group.schedule}</span>
          )}
          <span className="text-xs text-muted-foreground">· {group.member_count ?? 0} учеников</span>
          {group.max_members && (
            <span className="text-xs text-muted-foreground">/ {group.max_members}</span>
          )}
        </div>
        {group.price && (
          <span className="text-xs text-muted-foreground mt-0.5">{group.price} ₽</span>
        )}
      </div>
      <div className="flex gap-1.5 shrink-0" onClick={(e) => e.stopPropagation()}>
        <button
          onClick={() => onEdit(group)}
          className="rounded-xl p-2 hover:bg-muted/50 transition-colors"
        >
          <Edit3 className="h-4 w-4 text-muted-foreground" />
        </button>
        <button
          onClick={() => onDelete(group)}
          className="rounded-xl p-2 hover:bg-destructive/10 transition-colors"
        >
          <Trash2 className="h-4 w-4 text-destructive" />
        </button>
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
            className="w-full rounded-2xl border border-border/60 bg-white px-4 py-2.5 text-sm"
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
            className="w-full rounded-2xl border border-border/60 bg-white px-4 py-2.5 text-sm"
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
              className="w-full rounded-2xl border border-border/60 bg-white px-4 py-2.5 text-sm"
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
              className="w-full rounded-2xl border border-border/60 bg-white px-4 py-2.5 text-sm"
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
    <div>
      <ScreenHeader
        title="Группы"
        subtitle={groups?.length ? `${groups.length} групп` : undefined}
        action={
          <button
            onClick={() => setShowCreate(true)}
            className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-white"
          >
            <Plus className="h-5 w-5" />
          </button>
        }
      />

      <div className="px-5 pb-24 flex flex-col gap-4">
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
