'use client'

import { usePendingLeads, useApproveLead, useRejectLead } from '@/lib/hooks'
import { Card } from '@/components/ui/card'
import { SkeletonList } from '@/components/ui/skeleton'
import { ArrowLeft, Check, X, UserPlus } from 'lucide-react'
import Link from 'next/link'

function levelLabel(level: string): string {
  switch (level) {
    case 'beginner': return 'Начинающий'
    case 'intermediate': return 'Средний'
    case 'advanced': return 'Продвинутый'
    default: return level
  }
}

export default function LeadsPage() {
  const { data: leads, isLoading } = usePendingLeads()
  const approve = useApproveLead()
  const reject = useRejectLead()

  return (
    <div className="pb-24">
      <div className="px-5 pt-6 pb-2 flex items-center gap-3">
        <Link href="/dashboard/home" className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 shrink-0">
          <ArrowLeft className="h-4 w-4 text-primary" />
        </Link>
        <h1 className="text-heading font-bold text-foreground">Заявки</h1>
      </div>

      <div className="px-5 mt-4 flex flex-col gap-3">
        {isLoading && <SkeletonList count={3} />}

        {!isLoading && (!leads || leads.length === 0) && (
          <div className="text-center py-12">
            <div className="flex justify-center mb-3">
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10">
                <UserPlus className="h-6 w-6 text-primary" />
              </div>
            </div>
            <p className="text-base font-bold text-foreground">Новых заявок нет</p>
            <p className="text-sm text-muted-foreground mt-1">Новые заявки будут появляться здесь</p>
          </div>
        )}

        {leads?.map((lead) => (
          <Card key={lead.id} className="p-4">
            <div className="flex items-start justify-between mb-3">
              <div>
                <p className="text-base font-bold text-foreground">{lead.full_name}</p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {lead.reg_type === 'child' && lead.target_name
                    ? `Ребёнок: ${lead.target_name}`
                    : 'Запись себя'}
                </p>
              </div>
              <span className="text-[10px] font-bold uppercase tracking-wider text-orange-600 bg-orange-50 px-2 py-0.5 rounded-full">
                {levelLabel(lead.target_level)}
              </span>
            </div>
            {lead.phone && (
              <p className="text-sm text-muted-foreground mb-3">📞 {lead.phone}</p>
            )}
            <p className="text-[11px] text-muted-foreground mb-3">
              {new Date(lead.created_at).toLocaleString('ru')}
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => approve.mutate(lead.id)}
                disabled={approve.isPending}
                className="flex-1 flex items-center justify-center gap-1.5 rounded-xl bg-primary py-2.5 text-sm font-bold text-primary-foreground active:scale-95 transition-transform disabled:opacity-50"
              >
                <Check className="h-4 w-4" />
                Принять
              </button>
              <button
                onClick={() => reject.mutate(lead.id)}
                disabled={reject.isPending}
                className="flex-1 flex items-center justify-center gap-1.5 rounded-xl bg-destructive/10 py-2.5 text-sm font-bold text-destructive active:scale-95 transition-transform disabled:opacity-50"
              >
                <X className="h-4 w-4" />
                Отклонить
              </button>
            </div>
          </Card>
        ))}

        {approve.isSuccess && (
          <div className="rounded-2xl bg-emerald-50 p-4 text-sm font-semibold text-emerald-700 text-center">
            ✅ Заявка одобрена! Ученик создан.
          </div>
        )}
      </div>
    </div>
  )
}