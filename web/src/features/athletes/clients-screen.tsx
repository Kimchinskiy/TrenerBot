'use client'

import { useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useClients } from '@/lib/hooks'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { ScreenHeader, Spinner, Empty } from '@/components/ui/screen'
import { SkeletonList } from '@/components/ui/skeleton'
import { Search, Users } from 'lucide-react'
import type { Client } from '@/lib/types'

function ClientRow({ client, onClick }: { client: Client; onClick: () => void }) {
  const isActive = client.status === 'active'

  return (
    <Card
      className="flex items-center gap-4 py-4 px-5"
      onClick={onClick}
    >
      <Avatar className="h-12 w-12 border-2 border-white shadow-sm shrink-0">
        <AvatarFallback className="bg-primary/10 text-primary text-sm font-bold uppercase">
          {client.full_name.charAt(0)}
        </AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0">
        <p className="text-base font-bold text-foreground truncate">{client.full_name}</p>
        <div className="flex items-center gap-2 mt-0.5">
          {client.age && (
            <span className="text-xs text-muted-foreground">{client.age} лет</span>
          )}
          {client.subscription_ends_at && (
            <>
              <span className="text-border">·</span>
              <span className="text-xs text-muted-foreground">Абонемент</span>
            </>
          )}
        </div>
      </div>
      <div className="flex flex-col items-end gap-1 shrink-0">
        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider ${
          isActive
            ? 'bg-success-light text-success'
            : 'bg-muted text-muted-foreground'
        }`}>
          {isActive ? 'Активен' : client.status}
        </span>
      </div>
    </Card>
  )
}

export default function ClientsScreen() {
  const router = useRouter()
  const { data: clients, isLoading } = useClients()
  const [search, setSearch] = useState('')

  const filtered = useMemo(() => {
    if (!clients) return []
    if (!search.trim()) return clients
    const q = search.toLowerCase()
    return clients.filter(
      (c) =>
        c.full_name.toLowerCase().includes(q) ||
        c.phone?.toLowerCase().includes(q),
    )
  }, [clients, search])

  return (
    <div className="pb-24">
      <ScreenHeader title="Клиенты" subtitle={`${clients?.length || 0} человек`} />

      <div className="px-5 mb-4">
        <div className="relative">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/60" />
          <Input
            placeholder="Поиск клиента..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-10 bg-muted/30"
          />
        </div>
      </div>

      <div className="px-5">
        {isLoading && <SkeletonList count={5} />}

        {!isLoading && filtered.length === 0 && (
          <Empty
            text={search ? 'Никого не нашли' : 'Пока нет клиентов'}
            icon={<Users className="h-12 w-12" />}
          />
        )}

        <div className="flex flex-col gap-2">
          {filtered.map((client) => (
            <ClientRow
              key={client.id}
              client={client}
              onClick={() => router.push(`/dashboard/clients/${client.id}`)}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
