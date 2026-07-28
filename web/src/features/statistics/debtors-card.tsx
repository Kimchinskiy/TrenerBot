'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'
import { AlertTriangle, CheckCircle, ChevronRight } from 'lucide-react'
import type { DebtorsSummary } from '@/lib/types'

export function DebtorsCard({ debtors }: { debtors: DebtorsSummary }) {
  const [expanded, setExpanded] = useState(false)
  const hasDebtors = debtors.count > 0

  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.25, duration: 0.3 }}
      onClick={() => hasDebtors && setExpanded(!expanded)}
      className={`rounded-3xl bg-card border border-border/50 p-5 shadow-card transition-all duration-200 ${
        hasDebtors ? 'active:scale-[0.98] cursor-pointer' : ''
      }`}
    >
      {hasDebtors ? (
        <>
          <div className="flex items-center gap-3 mb-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-destructive/10 shrink-0">
              <AlertTriangle className="h-5 w-5 text-destructive" />
            </div>
            <div>
              <p className="text-sm font-bold text-foreground">Должники</p>
              <p className="text-xs text-muted-foreground">{debtors.count} клиентов</p>
            </div>
            <div className="ml-auto text-right">
              <p className="text-lg font-bold text-destructive">
                {debtors.total_debt.toLocaleString('ru-RU')} ₽
              </p>
            </div>
          </div>

          {expanded && (
            <div className="mt-3 pt-3 border-t border-border/40 flex flex-col gap-2">
              {debtors.items.map((item) => (
                <div key={item.client_id} className="flex items-center justify-between py-1">
                  <div>
                    <p className="text-sm font-semibold text-foreground">{item.full_name}</p>
                    {item.phone && (
                      <p className="text-xs text-muted-foreground">{item.phone}</p>
                    )}
                  </div>
                  <p className="text-sm font-bold text-destructive">
                    {item.debt.toLocaleString('ru-RU')} ₽
                  </p>
                </div>
              ))}
            </div>
          )}
        </>
      ) : (
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-success-light shrink-0">
            <CheckCircle className="h-5 w-5 text-success" />
          </div>
          <p className="text-sm font-bold text-foreground">Задолженностей нет</p>
        </div>
      )}
    </motion.div>
  )
}
