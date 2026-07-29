'use client'

import { useMe } from '@/lib/hooks'
import { ScreenHeader, Spinner, ErrorBox } from '@/components/ui/screen'
import { TelegramIntegrationCard } from '@/components/telegram-integration-card'

export default function SettingsPage() {
  const { data, isLoading, error } = useMe()

  if (isLoading) return <Spinner label="Загрузка..." />
  if (error) return <ErrorBox error={error} />

  return (
    <div className="pb-24 pt-6">
      <ScreenHeader title="Настройки" />
      <div className="px-5 flex flex-col gap-5">
        <TelegramIntegrationCard
          userId={data?.client?.user_id ?? undefined}
          hasBotAccess={data?.client?.bot_access}
          phone={data?.client?.phone ?? undefined}
        />
      </div>
    </div>
  )
}
