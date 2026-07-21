'use client'

import { useState } from 'react'
import { useGroups, useCreateGroup, useUpdateGroup, useDeleteGroup } from '@/lib/hooks'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ScreenHeader, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { Plus, Pencil, Trash2, Users } from 'lucide-react'
import type { Group } from '@/lib/types'

export default function GroupsScreen() {
  const { data: groups = [], isLoading, error } = useGroups()
  const createGroup = useCreateGroup()
  const updateGroup = useUpdateGroup()
  const deleteGroup = useDeleteGroup()

  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Group | null>(null)
  const [name, setName] = useState('')
  const [schedule, setSchedule] = useState('')
  const [location, setLocation] = useState('')
  const [maxMembers, setMaxMembers] = useState('')
  const [price, setPrice] = useState('')

  const reset = () => {
    setEditing(null)
    setName('')
    setSchedule('')
    setLocation('')
    setMaxMembers('')
    setPrice('')
  }

  const openCreate = () => {
    reset()
    setOpen(true)
  }

  const openEdit = (g: Group) => {
    setEditing(g)
    setName(g.name || '')
    setSchedule(g.schedule || '')
    setLocation(g.location || '')
    setMaxMembers(g.max_members != null ? String(g.max_members) : '')
    setPrice(g.price != null ? String(g.price) : '')
    setOpen(true)
  }

  const handleSubmit = async () => {
    const data = {
      name,
      schedule: schedule || undefined,
      location: location || undefined,
      max_members: maxMembers ? Number(maxMembers) : undefined,
      price: price ? Number(price) : undefined,
      active: 1,
    }
    if (editing) {
      await updateGroup.mutateAsync({ id: editing.id, ...data })
    } else {
      await createGroup.mutateAsync(data)
    }
    setOpen(false)
    reset()
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Удалить группу?')) return
    await deleteGroup.mutateAsync(id)
  }

  return (
    <div className="px-4 pb-24 pt-6">
      <ScreenHeader
        title="Группы"
        action={
          <Button size="icon" className="rounded-full" onClick={openCreate}>
            <Plus className="h-5 w-5" />
          </Button>
        }
      />

      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}
      {!isLoading && !error && groups.length === 0 && <Empty text="Группы пока не созданы" />}

      <div className="mt-4 flex flex-col gap-3">
        {groups.map((g) => (
          <Card key={g.id} className="p-4 shadow-sm border-border/80">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-bold text-foreground">{g.name || 'Без названия'}</h3>
                <p className="text-sm text-muted-foreground mt-1">
                  {g.schedule ? `📅 ${g.schedule}` : 'Расписание не указано'}
                </p>
                <p className="text-sm text-muted-foreground">
                  {g.location ? `📍 ${g.location}` : ''}
                  {g.max_members ? ` · 👥 до ${g.max_members}` : ''}
                </p>
                {g.price ? <p className="text-sm text-muted-foreground">💰 {g.price} ₽</p> : null}
              </div>
              <div className="flex gap-2">
                <Button size="icon" variant="ghost" onClick={() => openEdit(g)}>
                  <Pencil className="h-4 w-4" />
                </Button>
                <Button size="icon" variant="ghost" onClick={() => handleDelete(g.id)}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" onClick={() => setOpen(false)}>
          <div className="w-full max-w-lg rounded-2xl bg-background p-5 shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-lg font-bold mb-4">{editing ? 'Редактировать группу' : 'Новая группа'}</h2>
            <div className="flex flex-col gap-3">
              <div>
                <Label className="mb-1">Название</Label>
                <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Название группы" />
              </div>
              <div>
                <Label className="mb-1">Расписание</Label>
                <Input value={schedule} onChange={(e) => setSchedule(e.target.value)} placeholder="Пн, Ср, Пт 18:00" />
              </div>
              <div>
                <Label className="mb-1">Локация</Label>
                <Input value={location} onChange={(e) => setLocation(e.target.value)} placeholder="Зал 1" />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <Label className="mb-1">Макс. участников</Label>
                  <Input type="number" value={maxMembers} onChange={(e) => setMaxMembers(e.target.value)} placeholder="10" />
                </div>
                <div>
                  <Label className="mb-1">Цена (₽)</Label>
                  <Input type="number" value={price} onChange={(e) => setPrice(e.target.value)} placeholder="1500" />
                </div>
              </div>
              <Button onClick={handleSubmit} disabled={!name.trim() || createGroup.isPending || updateGroup.isPending}>
                {editing ? 'Сохранить' : 'Создать'}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
