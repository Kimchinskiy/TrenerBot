import { useState } from 'react'
import { useFaq, useSocialMedia } from '../lib/hooks'
import { ScreenHeader, Card, Spinner, Empty, ErrorBox } from '../components/ui'

export default function Social({ onBack }: { onBack: () => void }) {
  const { data: links, isLoading } = useSocialMedia()
  const [q, setQ] = useState('')
  const faq = useFaq(q)

  return (
    <div>
      <ScreenHeader title="Соцсети / FAQ" onBack={onBack} />
      <div className="px-4 pb-6">
        <div className="mb-2 text-sm text-tg-hint">Наши соцсети</div>
        {isLoading && <Spinner />}
        <div className="flex flex-col gap-2">
          {links &&
            Object.entries(links).map(([name, url]) => (
              <a key={name} href={url} target="_blank" rel="noreferrer">
                <Card className="flex items-center justify-between">
                  <span className="font-medium capitalize">{name}</span>
                  <span className="text-tg-link">→</span>
                </Card>
              </a>
            ))}
        </div>

        <div className="mb-2 mt-5 text-sm text-tg-hint">Частые вопросы</div>
        <input
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Спросите: цена, расписание, медицинские..."
          className="w-full rounded-xl bg-tg-secondary px-3 py-2 text-tg-text outline-none"
        />
        {faq.isError && <ErrorBox error={faq.error} />}
        {faq.data?.answer && (
          <Card className="mt-3 whitespace-pre-wrap text-sm">{faq.data.answer}</Card>
        )}
        {q && !faq.data?.answer && !faq.isFetching && <Empty text="Нет ответа" />}
      </div>
    </div>
  )
}
