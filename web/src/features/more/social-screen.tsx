'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/components/auth-provider'
import { useSocialMedia, useSaveSocialLinks } from '@/lib/hooks'
import { ScreenHeader, Card, Spinner } from '@/components/ui/screen'
import { Button } from '@/components/ui'
import { Switch } from '@/components/ui/switch'
import { Accordion } from '@/components/ui/accordion'
import { Save, Instagram, Globe, Youtube, MessageCircle, Phone, CircleDollarSign, Calendar, Heart, Shirt } from 'lucide-react'
import type { SocialLink } from '@/lib/types'

const FAQ_ITEMS = [
  {
    value: 'price',
    icon: <CircleDollarSign className="h-4 w-4" />,
    question: 'Стоимость занятий',
    answer: '💰 Стоимость занятий:\n• Разовое — 1500₽\n• Абонемент на 8 — 10000₽ (1250₽/зан)\n• Абонемент на 12 — 13200₽ (1100₽/зан)\n\nЕсть скидка 10% на первый абонемент!',
  },
  {
    value: 'schedule',
    icon: <Calendar className="h-4 w-4" />,
    question: 'Расписание тренировок',
    answer: '📅 Расписание:\n• Понедельник 09:00, 19:00\n• Среда 09:00, 19:00\n• Пятница 09:00, 18:00\n• Суббота 10:00\n\nТочное расписание на неделю — кнопка «Моё расписание»',
  },
  {
    value: 'medical',
    icon: <Heart className="h-4 w-4" />,
    question: 'Медицинские ограничения',
    answer: '🏥 Медицинские ограничения:\nПри наличии хронических заболеваний или травм — обязательно справка от врача. Тренер скорректирует нагрузку под ваши особенности.',
  },
  {
    value: 'gear',
    icon: <Shirt className="h-4 w-4" />,
    question: 'Что взять с собой',
    answer: '👟 Что взять:\n• Удобная спортивная одежда\n• Кроссовки с чистым подошвой\n• Вода\n• Полотенце (по желанию)\n\nДуш и раздевалки на месте.',
  },
]

const PLATFORM_META: Record<string, { label: string; icon: React.ReactNode; placeholder: string }> = {
  instagram: { label: 'Instagram', icon: <Instagram className="h-5 w-5" />, placeholder: 'https://instagram.com/...' },
  telegram: { label: 'Telegram', icon: <MessageCircle className="h-5 w-5" />, placeholder: 'https://t.me/...' },
  vk: { label: 'VK', icon: <Globe className="h-5 w-5" />, placeholder: 'https://vk.com/...' },
  youtube: { label: 'YouTube', icon: <Youtube className="h-5 w-5" />, placeholder: 'https://youtube.com/@...' },
  whatsapp: { label: 'WhatsApp', icon: <Phone className="h-5 w-5" />, placeholder: 'https://wa.me/...' },
}

const DEFAULT_PLATFORMS = ['instagram', 'telegram', 'vk', 'youtube', 'whatsapp']

export default function SocialScreen() {
  const router = useRouter()
  const { role } = useAuth()
  const { data: linksMap, isLoading } = useSocialMedia()
  const saveLinks = useSaveSocialLinks()
  const [links, setLinks] = useState<SocialLink[]>([])
  const isCoach = role === 'coach' || role === 'admin'

  useEffect(() => {
    if (linksMap) {
      const existing = DEFAULT_PLATFORMS.map((p) => ({
        platform: p,
        url: (linksMap as Record<string, string>)[p] || '',
        enabled: !!((linksMap as Record<string, string>)[p]),
      }))
      setLinks(existing)
    }
  }, [linksMap])

  const updateLink = (platform: string, field: 'url' | 'enabled', value: string | boolean) => {
    setLinks((prev) =>
      prev.map((l) => (l.platform === platform ? { ...l, [field]: value } : l))
    )
  }

  const handleSave = () => {
    saveLinks.mutate(links)
  }

  return (
    <div>
      <ScreenHeader title="Соцсети / FAQ" onBack={() => router.back()} />
      <div className="px-4 pb-24 flex flex-col gap-5">
        <div>
          <div className="mb-3 px-1 text-sm font-bold tracking-wider text-muted-foreground uppercase flex items-center justify-between">
            <span>Настройка соцсетей</span>
            {isCoach && (
              <Button
                onClick={handleSave}
                disabled={saveLinks.isPending}
                size="sm"
                className="h-8 text-xs font-bold shadow-md"
              >
                <Save className="h-3.5 w-3.5 mr-1" />
                {saveLinks.isPending ? 'Сохранение...' : 'Сохранить'}
              </Button>
            )}
          </div>
          {isLoading && <Spinner />}
          <div className="flex flex-col gap-2">
            {links.map((link) => {
              const meta = PLATFORM_META[link.platform]
              if (!meta) return null
              return (
                <Card key={link.platform} className="border-border/80 shadow-sm">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <span className="text-muted-foreground">{meta.icon}</span>
                      <span className="font-bold text-sm">{meta.label}</span>
                    </div>
                    {isCoach && (
                      <Switch
                        checked={link.enabled}
                        onCheckedChange={(v) => updateLink(link.platform, 'enabled', v)}
                      />
                    )}
                  </div>
                  {isCoach ? (
                    <input
                      className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm mt-1"
                      placeholder={meta.placeholder}
                      value={link.url || ''}
                      onChange={(e) => updateLink(link.platform, 'url', e.target.value)}
                    />
                  ) : link.enabled && link.url ? (
                    <a
                      href={link.url}
                      target="_blank"
                      rel="noreferrer"
                      className="text-sm text-primary underline mt-1 block"
                    >
                      {link.url}
                    </a>
                  ) : null}
                  {saveLinks.isSuccess && (
                    <p className="text-emerald-400 text-xs mt-1">Сохранено</p>
                  )}
                </Card>
              )
            })}
          </div>
        </div>

        <div>
          <div className="mb-2 px-1 text-sm font-bold tracking-wider text-muted-foreground uppercase">Частые вопросы</div>
          <Card className="border-border/80 shadow-sm overflow-hidden">
            <Accordion defaultValue={['price']}>
              {FAQ_ITEMS.map((item) => (
                <Accordion.Item key={item.value} value={item.value}>
                  <Accordion.Trigger>
                    <span className="flex items-center gap-2">
                      <span className="text-muted-foreground">{item.icon}</span>
                      {item.question}
                    </span>
                  </Accordion.Trigger>
                  <Accordion.Panel>{item.answer}</Accordion.Panel>
                </Accordion.Item>
              ))}
            </Accordion>
          </Card>
        </div>
      </div>
    </div>
  )
}
