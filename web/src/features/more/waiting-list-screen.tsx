'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useAddWaitingList, useRemoveWaitingList, useWaitingList } from '@/lib/hooks'
import { Card, ScreenHeader, Spinner, Empty, ErrorBox, Row } from '@/components/ui/screen'
import { Button, Input } from '@/components/ui'
import { haptics } from '@/services/telegram'

export default function WaitingListScreen() {
  const router = useRouter()
  const { data, isLoading, error } = useWaitingList()
  const remove = useRemoveWaitingList()
  const add = useAddWaitingList()
  const [clientId, setClientId] = useState('')

  const onAdd = async () => {
    const id = Number(clientId)
    if (!id) return
    try {
      await add.mutateAsync(id)
      haptics.success()
      setClientId('')
    } catch {
      haptics.error()
    }
  }

  return (
    <div>
      <ScreenHeader title="Лист ожидания" subtitle="Очередь на тренировки" onBack={() => router.back()} />
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}

      <div className="px-4 pb-4">
        <div className="flex gap-2.5">
          <Input
            value={clientId}
            onChange={(e) => setClientId(e.target.value.replace(/\D/g, ''))}
            placeholder="Введите ID клиента"
            inputMode="numeric"
            className="flex-1"
          />
          <Button
            onClick={onAdd}
            disabled={!clientId || add.isPending}
            className="h-12 w-12 shrink-0 font-extrabold text-xl flex items-center justify-center"
          >
            +
          </Button>
        </div>
      </div>

      {!isLoading && data && data.length === 0 && <Empty text="Лист ожидания пуст" />}
      <div className="flex flex-col gap-3 px-4 pb-24">
        {(data || []).map((w) => (
          <Card key={w.id} className="flex items-center justify-between border-border/80 shadow-sm relative overflow-hidden">
            <div className="absolute top-0 right-0 h-16 w-16 bg-primary/5 rounded-full blur-2xl" />
            <div className="flex-1 mr-4">
              <div className="text-base font-bold text-foreground">
                {w.position}. {w.name}
              </div>
              <div className="mt-1.5 flex flex-col gap-1">
                <Row label="Телефон" value={w.phone} />
              </div>
            </div>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => remove.mutate(w.id)}
              className="shrink-0 font-semibold h-8 px-3"
            >
              Убрать
            </Button>
          </Card>
        ))}
      </div>
    </div>
  )
}
