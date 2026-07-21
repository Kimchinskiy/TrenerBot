'use client'

import { useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useGroup, useGroupClients, useAddClientToGroup, useRemoveClientFromGroup, useClients } from '@/lib/hooks'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { ScreenHeader, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { ChevronLeft, Plus, UserMinus } from 'lucide-react'
import type { GroupMember } from '@/lib/types'

export default function GroupDetailScreen() {
  const params = useParams()
  const router = useRouter()
  const groupId = Number(params.id)
  const { data: group, isLoading, error } = useGroup(groupId)
  const { data: members = [], isLoading: membersLoading } = useGroupClients(groupId)
  const { data: allClients = [] } = useClients()
  const addClient = useAddClientToGroup()
  const removeClient = useRemoveClientFromGroup()

  const [open, setOpen] = useState(false)
  const [selectedClientId, setSelectedClientId] = useState<number | null>(null)

  const available = allClients.filter((c) => !members.some((m) => m.client_id === c.id))

  const handleAdd = async () => {
    if (!selectedClientId) return
    await addClient.mutateAsync({ groupId, clientId: selectedClientId, role: 'member' })
    setSelectedClientId(null)
    setOpen(false)
  }

  const handleRemove = async (clientId: number) => {
    await removeClient.mutateAsync({ groupId, clientId })
  }

  if (!isLoading && (!group || Number.isNaN(groupId))) {
    return (
      <div className="px-4 pb-24 pt-6">
        <ScreenHeader title="Группа не найдена" onBack={() => router.back()} />
        <Empty text="Группа не найдена или удалена" />
      </div>
    )
  }

  return (
    <div className="px-4 pb-24 pt-6">
      <ScreenHeader
        title={group?.name || 'Группа'}
        subtitle={group?.schedule || ''}
        onBack={() => router.back()}
        action={
          <Button size="icon" className="rounded-full" onClick={() => setOpen(true)}>
            <Plus className="h-5 w-5" />
          </Button>
        }
      />

      {(isLoading || membersLoading) && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}

      {!isLoading && !membersLoading && !error && (
        <>
          {group?.location && <p className="text-sm text-muted-foreground mb-4">📍 {group.location}</p>}
          {members.length === 0 && <Empty text="В группе пока нет участников" />}
          <div className="mt-4 flex flex-col gap-3">
            {members.map((m: GroupMember) => (
              <Card key={m.id} className="p-4 shadow-sm border-border/80 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center text-primary font-bold text-sm">
                    {(m.client_name || '?').charAt(0)}
                  </div>
                  <div>
                    <p className="font-medium text-foreground">{m.client_name || `Клиент #${m.client_id}`}</p>
                    <p className="text-xs text-muted-foreground">{m.role === 'assistant' ? 'Ассистент' : 'Участник'}</p>
                  </div>
                </div>
                <Button size="icon" variant="ghost" onClick={() => handleRemove(m.client_id)}>
                  <UserMinus className="h-4 w-4" />
                </Button>
              </Card>
            ))}
          </div>
        </>
      )}

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" onClick={() => setOpen(false)}>
          <div className="w-full max-w-lg rounded-2xl bg-background p-5 shadow-2xl" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-lg font-bold mb-3">Добавить клиента</h2>
            {available.length === 0 ? (
              <p className="text-sm text-muted-foreground">Нет доступных клиентов</p>
            ) : (
              <div className="flex flex-col gap-2 max-h-80 overflow-y-auto">
                {available.map((c) => (
                  <button
                    key={c.id}
                    onClick={() => setSelectedClientId(c.id)}
                    className={`w-full text-left rounded-xl border px-4 py-3 text-sm transition-colors ${selectedClientId === c.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/60'}`}
                  >
                    {c.full_name}
                  </button>
                ))}
              </div>
            )}
            <div className="mt-4 flex gap-2">
              <Button variant="outline" className="flex-1" onClick={() => setOpen(false)}>Отмена</Button>
              <Button className="flex-1" onClick={handleAdd} disabled={!selectedClientId || addClient.isPending}>Добавить</Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
