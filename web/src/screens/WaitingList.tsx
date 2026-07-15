import { useState } from 'react'
import { useAddWaitingList, useRemoveWaitingList, useWaitingList } from '../lib/hooks'
import { Card, ScreenHeader, Spinner, Button, Empty, ErrorBox, Row } from '../components/ui'
import { haptics } from '../lib/theme'

export default function WaitingList({ onBack }: { onBack: () => void }) {
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
      <ScreenHeader title="Лист ожидания" subtitle="Очередь на тренировки" onBack={onBack} />
      {isLoading && <Spinner label="Загрузка..." />}
      {error && <ErrorBox error={error} />}

      <div className="px-4 pb-3">
        <div className="mb-2 flex gap-2">
          <input
            value={clientId}
            onChange={(e) => setClientId(e.target.value.replace(/\D/g, ''))}
            placeholder="ID клиента"
            inputMode="numeric"
            className="flex-1 rounded-xl bg-tg-secondary px-3 py-2 text-tg-text outline-none"
          />
          <button
            onClick={onAdd}
            disabled={!clientId || add.isPending}
            className="rounded-xl bg-tg-button px-4 py-2 text-tg-button-text disabled:opacity-50"
          >
            +
          </button>
        </div>
      </div>

      {!isLoading && data && data.length === 0 && <Empty text="Лист ожидания пуст" />}
      <div className="flex flex-col gap-2 px-4 pb-24">
        {(data || []).map((w) => (
          <Card key={w.id} className="flex items-center justify-between">
            <div>
              <div className="font-semibold">
                {w.position}. {w.name}
              </div>
              <Row label="Телефон" value={w.phone} />
            </div>
            <button
              onClick={() => remove.mutate(w.id)}
              className="rounded-lg bg-red-600/80 px-3 py-1 text-sm text-white"
            >
              Убрать
            </button>
          </Card>
        ))}
      </div>
    </div>
  )
}
