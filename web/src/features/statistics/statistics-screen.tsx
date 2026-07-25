'use client'

import { useState } from 'react'
import { useStatistics } from '@/lib/hooks'
import { Spinner } from '@/components/ui/screen'
import { StatisticsHeader } from './statistics-header'
import { PeriodSwitcher } from './period-switcher'
import { StatisticsCards } from './statistics-cards'
import { DebtorsCard } from './debtors-card'
import { IncomeChart } from './income-chart'
import { QuickOverview } from './quick-overview'

export function StatisticsScreen() {
  const [period, setPeriod] = useState('week')
  const { data, isLoading } = useStatistics(period)

  return (
    <div className="flex flex-col gap-5 px-5 pb-6">
      <StatisticsHeader />
      <PeriodSwitcher value={period} onChange={setPeriod} />

      {isLoading ? (
        <Spinner label="Загрузка статистики..." />
      ) : !data ? (
        <p className="text-sm text-muted-foreground text-center py-10">
          Не удалось загрузить статистику
        </p>
      ) : (
        <>
          <StatisticsCards stats={data} />
          <DebtorsCard debtors={data.debtors} />
          <IncomeChart data={data.income_chart} period={data.period} />
          <QuickOverview overview={data.quick_overview} />
        </>
      )}
    </div>
  )
}
