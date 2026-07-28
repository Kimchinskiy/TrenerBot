'use client'

import { useState } from 'react'
import { useLinkChild, useMe } from '@/lib/hooks'
import { Card, Spinner, ErrorBox } from '@/components/ui/screen'
import { Button } from '@/components/ui'
import { UserPlus, Check } from 'lucide-react'

export default function LinkChildSection() {
  const { data: me, isLoading } = useMe()
  const link = useLinkChild()
  const [fullName, setFullName] = useState('')
  const [birthDate, setBirthDate] = useState('')

  if (isLoading) return <Spinner label="Загрузка..." />
  if (!me || me.role !== 'parent') return null

  const alreadyHasChildren = me.children && me.children.length > 0

  return (
    <Card className="mb-4 shadow-md border-border/80">
      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          <UserPlus className="h-5 w-5 text-primary" />
          <h3 className="font-bold text-base">
            {alreadyHasChildren ? 'Привязать ещё одного ребёнка' : 'Привязать ребёнка'}
          </h3>
        </div>
        <p className="text-xs text-muted-foreground">
          Найдите ребёнка по имени и дате рождения
        </p>
        <input
          className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-3 text-sm"
          placeholder="Имя и фамилия ребёнка"
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
        />
        <input
          type="date"
          className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-3 text-sm"
          value={birthDate}
          onChange={(e) => setBirthDate(e.target.value)}
        />
        <Button
          onClick={() => {
            if (!fullName.trim() || !birthDate) return
            link.mutate({ fullName: fullName.trim(), birthDate })
          }}
          disabled={link.isPending || !fullName.trim() || !birthDate}
          className="w-full font-bold shadow-md"
        >
          {link.isPending ? 'Поиск...' : 'Найти и привязать'}
        </Button>
        {link.isSuccess && (
          <div className="flex items-center gap-2 text-success text-sm font-bold">
            <Check className="h-4 w-4" />
            <span>Ребёнок {link.data?.child_name} привязан!</span>
          </div>
        )}
        {link.isError && (
          <div className="text-destructive text-sm font-semibold">{link.error?.message}</div>
        )}
      </div>
    </Card>
  )
}
