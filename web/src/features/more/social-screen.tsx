'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useFaq, useSocialMedia } from '@/lib/hooks'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox } from '@/components/ui/screen'
import { Input } from '@/components/ui'
import { ChevronRight } from 'lucide-react'

export default function SocialScreen() {
  const router = useRouter()
  const { data: links, isLoading } = useSocialMedia()
  const [q, setQ] = useState('')
  const faq = useFaq(q)

  return (
    <div>
      <ScreenHeader title="Соцсети / FAQ" onBack={() => router.back()} />
      <div className="px-4 pb-24 flex flex-col gap-5">
        <div>
          <div className="mb-3 px-1 text-sm font-bold tracking-wider text-muted-foreground uppercase">Наши соцсети</div>
          {isLoading && <Spinner />}
          <div className="flex flex-col gap-3">
            {links &&
              Object.entries(links).map(([name, url]) => (
                <a key={name} href={url as string} target="_blank" rel="noreferrer" className="block">
                  <Card className="flex items-center justify-between py-4 px-5 border-border/80 shadow-sm hover:border-foreground/10 transition-all duration-200">
                    <span className="font-bold text-foreground capitalize">{name}</span>
                    <ChevronRight className="h-5 w-5 text-muted-foreground/80" />
                  </Card>
                </a>
              ))}
          </div>
        </div>

        <div>
          <div className="mb-3 px-1 text-sm font-bold tracking-wider text-muted-foreground uppercase">Частые вопросы</div>
          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Спросите: цена, расписание..."
            className="w-full"
          />
          {faq.isError && <ErrorBox error={faq.error} />}
          {faq.data?.answer && (
            <Card className="mt-4 border-border bg-card p-5 shadow-sm text-sm text-foreground/90 whitespace-pre-wrap leading-relaxed relative overflow-hidden">
               <div className="absolute top-0 right-0 h-16 w-16 bg-primary/5 rounded-full blur-2xl" />
               {faq.data.answer}
            </Card>
          )}
          {q && !faq.data?.answer && !faq.isFetching && <Empty text="Нет соответствий в базе знаний" />}
        </div>
      </div>
    </div>
  )
}
