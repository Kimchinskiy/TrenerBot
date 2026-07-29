'use client'

import { useState } from 'react'
import { useMe } from '@/lib/hooks'
import { api } from '@/lib/api'
import { ScreenHeader, Card, Spinner, ErrorBox, Row } from '@/components/ui/screen'
import { TelegramLoginButton } from '@/integrations/telegram/login-widget'
import { Button } from '@/components/ui/button'
import { CheckCircle2, XCircle, ExternalLink } from 'lucide-react'

const BOT_USERNAME = process.env.NEXT_PUBLIC_TELEGRAM_BOT || ''

export default function Settings() {
  const { data, isLoading, error, refetch } = useMe()
  const [linkError, setLinkError] = useState('')
  const [linkOk, setLinkOk] = useState(false)

  const handleTelegramAuth = async (fields: Record<string, string>) => {
    setLinkError('')
    setLinkOk(false)
    try {
      await api.linkTelegramWidget(fields)
      setLinkOk(true)
      refetch()
    } catch (e: any) {
      setLinkError(e?.message || 'Ошибка привязки')
    }
  }

  if (isLoading) return <Spinner label="Загрузка..." />
  if (error) return <ErrorBox error={error} />

  const telegramId = data?.telegram_id
  const bindLink = data?.user_id ? `https://t.me/${BOT_USERNAME}?start=bind_${data.user_id}` : ''

  return (
    <div className="pb-24 pt-6">
      <ScreenHeader title="Настройки" />

      <div className="px-5 flex flex-col gap-5 mt-5">
        {/* Telegram connection */}
        <Card className="p-4">
          <h3 className="text-base font-bold text-foreground mb-1">Telegram</h3>
          <p className="text-sm text-muted-foreground mb-4">
            Подключите Telegram, чтобы бот узнавал вас как тренера и вы могли управлять клиентами через бот.
          </p>

          {telegramId ? (
            <div className="flex items-center gap-3 p-3 rounded-2xl bg-success-light/20 border border-success/20">
              <CheckCircle2 className="h-5 w-5 text-success shrink-0" />
              <div>
                <p className="text-sm font-semibold text-foreground">Telegram подключён</p>
                <p className="text-xs text-muted-foreground">ID: {telegramId}</p>
              </div>
            </div>
          ) : (
            <div className="flex flex-col gap-4">
              <div className="flex items-center gap-3 p-3 rounded-2xl bg-muted/30 border border-border/30">
                <XCircle className="h-5 w-5 text-muted-foreground shrink-0" />
                <p className="text-sm text-muted-foreground">Telegram не подключён</p>
              </div>

              <div className="p-4 rounded-2xl bg-primary/5 border border-primary/20">
                <h4 className="text-sm font-bold text-foreground mb-2">Быстрая привязка через бота</h4>
                <p className="text-xs text-muted-foreground mb-3">
                  Нажмите на кнопку ниже — откроется Telegram с ботом @{BOT_USERNAME}. Нажмите «Начать» — привязка произойдёт автоматически, без регистрации.
                </p>
                <a
                  href={bindLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-center gap-2 w-full rounded-2xl bg-primary text-primary-foreground px-4 py-3 text-sm font-semibold hover:bg-primary/90 transition-colors"
                >
                  <ExternalLink className="h-4 w-4" />
                  Привязать Telegram
                </a>
              </div>

              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-border/40" />
                </div>
                <div className="relative flex justify-center text-xs">
                  <span className="bg-card px-2 text-muted-foreground">или через виджет</span>
                </div>
              </div>

              <TelegramLoginButton onAuth={handleTelegramAuth} />
              {linkOk && (
                <p className="text-sm text-success font-semibold text-center">Telegram успешно привязан!</p>
              )}
              {linkError && (
                <p className="text-sm text-destructive font-semibold text-center">{linkError}</p>
              )}
            </div>
          )}
        </Card>

        {/* App info */}
        <Card className="p-4">
          <Row label="Ваша роль" value={data?.role === 'coach' ? 'Тренер' : data?.role === 'admin' ? 'Администратор' : data?.role === 'parent' ? 'Родитель' : 'Клиент'} />
          <Row label="Версия приложения" value="1.0.0" />
        </Card>
      </div>
    </div>
  )
}
