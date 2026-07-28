'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { ScreenHeader, Card, Empty, Spinner } from '@/components/ui/screen'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useClients, useGroups, useNotificationsPreview, useSendNotification } from '@/lib/hooks'
import { haptics } from '@/services/telegram'
import {
  Send,
  Bell,
  Users,
  Calendar,
  UserCheck,
  CheckCircle2,
  AlertCircle,
  Search,
  Check,
} from 'lucide-react'

type RecipientFilter = 'all' | 'group' | 'today' | 'tomorrow' | 'manual'

export default function NotificationsScreen() {
  const router = useRouter()
  const { data: clients } = useClients()
  const { data: groups } = useGroups()

  const [filter, setFilter] = useState<RecipientFilter>('all')
  const [selectedGroupId, setSelectedGroupId] = useState<number | undefined>(undefined)
  const [selectedClientIds, setSelectedClientIds] = useState<number[]>([])

  const [clientSearch, setClientSearch] = useState('')

  const [title, setTitle] = useState('')
  const [text, setText] = useState('')
  const [sentMessage, setSentMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const { data: previewData, isLoading: isPreviewLoading } = useNotificationsPreview(
    filter,
    filter === 'group' ? selectedGroupId : undefined,
    filter === 'manual' ? selectedClientIds : undefined,
  )

  const sendNotification = useSendNotification()

  const toggleClientSelection = (id: number) => {
    setSelectedClientIds((prev) =>
      prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id],
    )
  }

  const selectAllFilteredClients = (filteredIds: number[]) => {
    const allSelected = filteredIds.every((id) => selectedClientIds.includes(id))
    if (allSelected) {
      setSelectedClientIds((prev) => prev.filter((id) => !filteredIds.includes(id)))
    } else {
      setSelectedClientIds((prev) => Array.from(new Set([...prev, ...filteredIds])))
    }
  }

  const handleSend = () => {
    if (!title.trim() || !text.trim()) return
    setError(null)
    setSentMessage(null)

    sendNotification.mutate(
      {
        filter,
        group_id: filter === 'group' ? selectedGroupId : undefined,
        client_ids: filter === 'manual' ? selectedClientIds : undefined,
        title: title.trim(),
        text: text.trim(),
      },
      {
        onSuccess: (res) => {
          haptics.success()
          setSentMessage(`Оповещение успешно отправлено! Поставлено в очередь: ${res.enqueued}, пропущено: ${res.skipped}.`)
          setTitle('')
          setText('')
          setTimeout(() => setSentMessage(null), 5000)
        },
        onError: (err: any) => {
          haptics.error()
          setError(err?.message || 'Ошибка при отправке оповещения')
        },
      },
    )
  }

  const filteredClientsList = clients?.filter((c) =>
    c.full_name.toLowerCase().includes(clientSearch.toLowerCase()),
  )

  const recipientCount = previewData?.total ?? 0

  return (
    <div>
      <ScreenHeader title="Оповещения" onBack={() => router.back()} />

      <div className="px-5 flex flex-col gap-5 pb-24">
        {/* Recipient Filter Selection */}
        <section>
          <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">
            Получатели оповещения
          </h3>
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
            {[
              { key: 'all', label: 'Все клиенты', icon: Bell },
              { key: 'group', label: 'По группе', icon: Users },
              { key: 'today', label: 'Занимающиеся сегодня', icon: Calendar },
              { key: 'tomorrow', label: 'Занимающиеся завтра', icon: Calendar },
              { key: 'manual', label: 'Выбор вручную', icon: UserCheck },
            ].map(({ key, label, icon: Icon }) => (
              <button
                key={key}
                onClick={() => {
                  setFilter(key as RecipientFilter)
                  if (key === 'group' && !selectedGroupId && groups && groups.length > 0) {
                    setSelectedGroupId(groups[0].id)
                  }
                }}
                className={`flex items-center gap-2 rounded-2xl p-3 text-xs font-semibold transition-all border text-left ${
                  filter === key
                    ? 'bg-primary text-primary-foreground border-primary shadow-sm'
                    : 'bg-card text-foreground border-border/60 hover:bg-muted/60 shadow-card'
                }`}
              >
                <Icon className="h-4 w-4 shrink-0" />
                <span className="truncate">{label}</span>
              </button>
            ))}
          </div>
        </section>

        {/* Group Selector Dropdown */}
        {filter === 'group' && (
          <section className="animate-in fade-in duration-200">
            <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1.5 block px-1">
              Выберите группу
            </label>
            <select
              value={selectedGroupId || ''}
              onChange={(e) => setSelectedGroupId(Number(e.target.value))}
              className="w-full rounded-2xl border border-border/60 bg-card text-foreground px-4 py-3 text-sm font-medium focus:outline-none focus:ring-2 focus:ring-primary/20 shadow-sm"
            >
              {groups?.map((g) => (
                <option key={g.id} value={g.id} className="bg-card text-foreground">
                  {g.name || `Группа #${g.id}`} ({g.member_count || 0} уч.)
                </option>
              ))}
            </select>
          </section>
        )}

        {/* Manual Client Multi-Selection */}
        {filter === 'manual' && (
          <section className="flex flex-col gap-3 animate-in fade-in duration-200">
            <div className="flex items-center justify-between px-1">
              <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                Выберите клиентов ({selectedClientIds.length})
              </label>
              {filteredClientsList && filteredClientsList.length > 0 && (
                <button
                  onClick={() => selectAllFilteredClients(filteredClientsList.map((c) => c.id))}
                  className="text-xs font-bold text-primary hover:underline"
                >
                  {filteredClientsList.every((c) => selectedClientIds.includes(c.id))
                    ? 'Снять все'
                    : 'Выбрать все'}
                </button>
              )}
            </div>

            <div className="relative">
              <Search className="absolute left-3.5 top-3 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Поиск по имени клиента..."
                value={clientSearch}
                onChange={(e) => setClientSearch(e.target.value)}
                className="w-full rounded-2xl border border-border/60 bg-card text-foreground placeholder:text-muted-foreground/60 pl-10 pr-4 py-2.5 text-sm shadow-sm"
              />
            </div>

            <Card className="max-h-60 overflow-y-auto !p-1 divide-y divide-border/30">
              {filteredClientsList && filteredClientsList.length > 0 ? (
                filteredClientsList.map((c) => {
                  const isSelected = selectedClientIds.includes(c.id)
                  return (
                    <div
                      key={c.id}
                      onClick={() => toggleClientSelection(c.id)}
                      className={`flex items-center justify-between p-3 cursor-pointer rounded-xl transition-colors ${
                        isSelected ? 'bg-primary/10' : 'hover:bg-muted/40'
                      }`}
                    >
                      <div className="flex flex-col">
                        <span className="text-sm font-semibold text-foreground">{c.full_name}</span>
                        <span className="text-xs text-muted-foreground">
                          {c.phone || 'Нет телефона'} {c.bot_access ? '· TG подключён' : ''}
                        </span>
                      </div>
                      <div
                        className={`flex h-5 w-5 items-center justify-center rounded-lg border transition-colors ${
                          isSelected
                            ? 'bg-primary text-primary-foreground border-primary'
                            : 'border-border/60 bg-card text-foreground'
                        }`}
                      >
                        {isSelected && <Check className="h-3.5 w-3.5" />}
                      </div>
                    </div>
                  )
                })
              ) : (
                <div className="p-4 text-center text-xs text-muted-foreground">
                  Клиенты не найдены
                </div>
              )}
            </Card>
          </section>
        )}

        {/* Recipients Preview Summary */}
        <section>
          <Card className="flex items-center justify-between bg-primary/10 border-primary/20">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/20 text-primary">
                <Users className="h-5 w-5" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">Будет отправлено</p>
                <p className="text-sm font-bold text-foreground">
                  {isPreviewLoading ? (
                    'Расчёт получателей...'
                  ) : (
                    `${recipientCount} получател${
                      recipientCount % 10 === 1 && recipientCount % 100 !== 11
                        ? 'ю'
                        : recipientCount % 10 >= 2 && recipientCount % 10 <= 4 && (recipientCount % 100 < 10 || recipientCount % 100 >= 20)
                        ? 'ям'
                        : 'ям'
                    }`
                  )}
                </p>
              </div>
            </div>
            {previewData?.recipients && previewData.recipients.length > 0 && (
              <span className="text-xs font-semibold px-3 py-1 bg-card rounded-full text-primary border border-primary/20 shadow-sm">
                {previewData.recipients.filter((r) => r.user_id !== null).length} в Telegram
              </span>
            )}
          </Card>
        </section>

        {/* Message Form */}
        <section>
          <h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-2 px-1">
            Сообщение
          </h3>
          <Card className="flex flex-col gap-3">
            <div>
              <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1 block">
                Заголовок
              </label>
              <Input
                placeholder="Заголовок оповещения (например: Внимание! Изменение расписания)"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
              />
            </div>

            <div>
              <label className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-1 block">
                Текст сообщения
              </label>
              <textarea
                rows={4}
                placeholder="Текст рассылки для клиентов..."
                value={text}
                onChange={(e) => setText(e.target.value)}
                className="w-full rounded-2xl border border-border/60 bg-card text-foreground placeholder:text-muted-foreground/60 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 resize-none shadow-sm transition-all"
              />
            </div>

            {sentMessage && (
              <div className="flex items-center gap-2 p-3 rounded-2xl bg-success-light text-success text-xs font-semibold">
                <CheckCircle2 className="h-4 w-4 shrink-0" />
                <span>{sentMessage}</span>
              </div>
            )}

            {error && (
              <div className="flex items-center gap-2 p-3 rounded-2xl bg-destructive/10 text-destructive text-xs font-semibold">
                <AlertCircle className="h-4 w-4 shrink-0" />
                <span>{error}</span>
              </div>
            )}

            <Button
              onClick={handleSend}
              disabled={
                !title.trim() ||
                !text.trim() ||
                sendNotification.isPending ||
                (filter === 'manual' && selectedClientIds.length === 0) ||
                (filter === 'group' && !selectedGroupId)
              }
              variant="gradient"
              size="lg"
              className="mt-1 rounded-2xl"
            >
              <Send className="h-4 w-4 mr-2" />
              {sendNotification.isPending
                ? 'Отправка...'
                : `Отправить ${recipientCount} получателям`}
            </Button>
          </Card>
        </section>
      </div>
    </div>
  )
}
