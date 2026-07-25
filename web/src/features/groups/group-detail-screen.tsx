'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import {
  useGroup,
  useGroupClients,
  useGroupAvailableClients,
  useAddClientToGroup,
  useRemoveClientFromGroup,
} from '@/lib/hooks'
import { ScreenHeader, Card, Spinner, ErrorBox, Empty } from '@/components/ui/screen'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  Plus,
  X,
  Search,
  UserPlus,
  MessageCircle,
  ExternalLink,
  Trash2,
  Users,
  Calendar,
  MoreVertical,
  Check,
} from 'lucide-react'
import type { Client, GroupMember } from '@/lib/types'

export default function GroupDetailScreen({ params }: { params: { id: string } }) {
  const groupId = Number(params.id)
  const router = useRouter()
  const { data: group, isLoading, error } = useGroup(groupId)
  const { data: members, isLoading: membersLoading } = useGroupClients(groupId)
  const { data: available, isLoading: availLoading } = useGroupAvailableClients(groupId)
  const addMutation = useAddClientToGroup()
  const removeMutation = useRemoveClientFromGroup()

  const [showAdd, setShowAdd] = useState(false)
  const [search, setSearch] = useState('')
  const [doneMsg, setDoneMsg] = useState('')
  const [menuClientId, setMenuClientId] = useState<number | null>(null)

  const showDone = (msg: string) => {
    setDoneMsg(msg)
    setTimeout(() => setDoneMsg(''), 3000)
  }

  const handleAdd = (clientId: number) => {
    addMutation.mutate(
      { groupId, clientId },
      {
        onSuccess: () => {
          setShowAdd(false)
          setSearch('')
          showDone('Ученик добавлен')
        },
      },
    )
  }

  const handleRemove = (clientId: number) => {
    if (!confirm('Удалить ученика из группы?')) return
    removeMutation.mutate(
      { groupId, clientId },
      {
        onSuccess: () => {
          setMenuClientId(null)
          showDone('Ученик удалён')
        },
      },
    )
  }

  const filteredAvailable = (available || []).filter((c) =>
    c.full_name?.toLowerCase().includes(search.toLowerCase()),
  )

  if (isLoading) return <Spinner label="Загрузка группы..." />
  if (error) return <ErrorBox error={error} />
  if (!group) return <ErrorBox error={new Error('Группа не найдена')} />

  return (
    <div>
      <ScreenHeader
        title={group.name || 'Группа'}
        subtitle={
          group.schedule
            ? `${group.schedule} · ${group.member_count ?? 0} учеников`
            : `${group.member_count ?? 0} учеников`
        }
        onBack={() => router.push('/dashboard/groups')}
        action={
          <button
            onClick={() => setShowAdd(true)}
            className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-white"
          >
            <UserPlus className="h-5 w-5" />
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

        {/* Group info */}
        <Card className="p-5">
          <div className="flex items-center gap-4 mb-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 shrink-0">
              <Users className="h-5 w-5 text-primary" />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-base font-bold text-foreground">{group.name || 'Без названия'}</h3>
              <p className="text-sm text-muted-foreground">{group.member_count ?? 0} учеников</p>
            </div>
          </div>
          <div className="flex flex-wrap gap-3 text-sm text-muted-foreground">
            {group.schedule && (
              <div className="flex items-center gap-1.5">
                <Calendar className="h-4 w-4" />
                <span>{group.schedule}</span>
              </div>
            )}
            {group.price && (
              <span className="font-semibold text-foreground">{group.price} ₽</span>
            )}
            {group.max_members && (
              <span className="text-muted-foreground">
                {group.member_count}/{group.max_members} мест
              </span>
            )}
            {group.location && (
              <span className="text-muted-foreground">📍 {group.location}</span>
            )}
          </div>
        </Card>

        {/* Add student modal */}
        {showAdd && (
          <Card className="p-5">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-bold text-foreground">Добавить ученика</h3>
              <button onClick={() => setShowAdd(false)} className="p-1 hover:bg-muted/50 rounded-lg transition-colors">
                <X className="h-4 w-4 text-muted-foreground" />
              </button>
            </div>
            <div className="relative mb-3">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Поиск по ФИО..."
                className="w-full rounded-2xl border border-border/60 bg-white pl-9 pr-4 py-2.5 text-sm"
              />
            </div>
            <div className="flex flex-col gap-1 max-h-64 overflow-y-auto">
              {availLoading && <Spinner label="Загрузка..." />}
              {!availLoading && filteredAvailable.length === 0 && (
                <p className="text-sm text-muted-foreground text-center py-4">
                  {search ? 'Ничего не найдено' : 'Нет доступных учеников'}
                </p>
              )}
              {filteredAvailable.map((c) => (
                <div
                  key={c.id}
                  className="flex items-center gap-3 p-2.5 rounded-xl hover:bg-muted/30 transition-colors cursor-pointer"
                  onClick={() => handleAdd(c.id)}
                >
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>{c.full_name?.[0] || '?'}</AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-semibold text-foreground truncate">{c.full_name}</p>
                    {c.phone && <p className="text-xs text-muted-foreground">{c.phone}</p>}
                  </div>
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 shrink-0">
                    <Plus className="h-4 w-4 text-primary" />
                  </div>
                </div>
              ))}
            </div>
          </Card>
        )}

        {/* Students list */}
        <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground px-1">
          Ученики
        </h3>

        {membersLoading && <Spinner label="Загрузка учеников..." />}
        {!membersLoading && (!members || members.length === 0) && (
          <Empty text="В группе пока нет учеников" icon={<Users className="h-8 w-8" />} />
        )}

        {members && members.length > 0 && (
          <div className="flex flex-col gap-1">
            {members.map((m) => (
              <StudentRow
                key={m.client_id}
                member={m}
                onRemove={() => handleRemove(m.client_id)}
                onViewCard={() => router.push(`/dashboard/clients/${m.client_id}`)}
                menuOpen={menuClientId === m.client_id}
                onToggleMenu={() =>
                  setMenuClientId(menuClientId === m.client_id ? null : m.client_id)
                }
                onCloseMenu={() => setMenuClientId(null)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function StudentRow({
  member,
  onRemove,
  onViewCard,
  menuOpen,
  onToggleMenu,
  onCloseMenu,
}: {
  member: GroupMember
  onRemove: () => void
  onViewCard: () => void
  menuOpen: boolean
  onToggleMenu: () => void
  onCloseMenu: () => void
}) {
  return (
    <Card className="flex items-center gap-3 relative">
      <Avatar className="h-9 w-9">
        <AvatarFallback>{member.client_name?.[0] || '?'}</AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-foreground truncate">
          {member.client_name || 'Без имени'}
        </p>
        <p className="text-xs text-muted-foreground">В группе с {member.joined_at?.slice(0, 10)}</p>
      </div>
      <div className="relative">
        <button
          onClick={onToggleMenu}
          className="rounded-xl p-2 hover:bg-muted/50 transition-colors"
        >
          <MoreVertical className="h-4 w-4 text-muted-foreground" />
        </button>
        {menuOpen && (
          <>
            <div className="fixed inset-0 z-10" onClick={onCloseMenu} />
            <div className="absolute right-0 top-full mt-1 z-20 min-w-[200px] rounded-2xl border border-border/60 bg-white p-1.5 shadow-lg">
              <button
                onClick={() => {
                  onCloseMenu()
                  onViewCard()
                }}
                className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-foreground hover:bg-muted/50 transition-colors"
              >
                <ExternalLink className="h-4 w-4 text-muted-foreground" />
                <span>Карточка ученика</span>
              </button>
              <button
                onClick={() => {
                  onCloseMenu()
                }}
                className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-foreground hover:bg-muted/50 transition-colors"
              >
                <MessageCircle className="h-4 w-4 text-muted-foreground" />
                <span>Написать сообщение</span>
              </button>
              <button
                onClick={() => {
                  onCloseMenu()
                }}
                className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-foreground hover:bg-muted/50 transition-colors"
              >
                <Users className="h-4 w-4 text-muted-foreground" />
                <span>Перевести в группу</span>
              </button>
              <hr className="my-1 border-border/40" />
              <button
                onClick={() => {
                  onCloseMenu()
                  onRemove()
                }}
                className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm text-destructive hover:bg-destructive/10 transition-colors"
              >
                <Trash2 className="h-4 w-4" />
                <span>Удалить из группы</span>
              </button>
            </div>
          </>
        )}
      </div>
    </Card>
  )
}
