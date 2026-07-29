'use client'

import { useState } from 'react'
import { useMe } from '@/lib/hooks'
import { api } from '@/lib/api'
import { ScreenHeader, Card, Spinner, ErrorBox, Row } from '@/components/ui/screen'
import { TelegramLoginButton } from '@/integrations/telegram/login-widget'
import { Button } from '@/components/ui/button'
import { CheckCircle2, XCircle, Smartphone } from 'lucide-react'

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
            <div className="flex flex-col gap-3">
              <div className="flex items-center gap-3 p-3 rounded-2xl bg-muted/30 border border-border/30">
                <XCircle className="h-5 w-5 text-muted-foreground shrink-0" />
                <p className="text-sm text-muted-foreground">Telegram не подключён</p>
              </div>
              <p className="text-xs text-muted-foreground">
                Нажмите на кнопку ниже, чтобы войти через Telegram. Бот @{process.env.NEXT_PUBLIC_TELEGRAM_BOT} будет привязан к вашему аккаунту.
              </p>
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
