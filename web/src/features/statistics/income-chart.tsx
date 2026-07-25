'use client'

import { motion } from 'framer-motion'
import type { ChartPoint } from '@/lib/types'

export function IncomeChart({ data, period }: { data: ChartPoint[]; period: string }) {
  if (!data || data.length === 0) {
    return (
      <div className="rounded-3xl bg-white p-5 shadow-card">
        <p className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-1">
          Доход
        </p>
        <p className="text-xs text-muted-foreground">Нет данных за выбранный период</p>
      </div>
    )
  }

  const values = data.map((d) => d.value)
  const maxVal = Math.max(...values, 1)
  const chartHeight = 140
  const barWidth = Math.max(12, Math.min(40, (280 - data.length) / data.length))

  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.3, duration: 0.3 }}
      className="rounded-3xl bg-white p-5 shadow-card"
    >
      <p className="text-sm font-bold uppercase tracking-wider text-muted-foreground mb-4">
        Доход
      </p>

      <div className="relative" style={{ height: chartHeight }}>
        <svg
          width="100%"
          height={chartHeight}
          viewBox={`0 0 ${data.length * (barWidth + 6)} ${chartHeight}`}
          className="overflow-visible"
        >
          <defs>
            <linearGradient id="chartGrad" x1="0" x2="0" y1="0" y2="1">
              <stop offset="0%" stopColor="hsl(168, 76%, 42%)" stopOpacity="0.3" />
              <stop offset="100%" stopColor="hsl(168, 76%, 42%)" stopOpacity="0.02" />
            </linearGradient>
          </defs>

          {/* Area fill */}
          <motion.path
            initial={{ d: `M 0,${chartHeight} L ${data.length * (barWidth + 6)},${chartHeight}` }}
            animate={{
              d: `M 0,${chartHeight} ${data
                .map(
                  (d, i) =>
                    `L ${i * (barWidth + 6) + barWidth / 2},${chartHeight - (d.value / maxVal) * chartHeight * 0.85}`,
                )
                .join(' ')} L ${(data.length - 1) * (barWidth + 6) + barWidth / 2},${chartHeight} Z`,
            }}
            transition={{ duration: 0.5, ease: 'easeOut' }}
            fill="url(#chartGrad)"
          />

          {/* Line */}
          <motion.path
            initial={{ d: '' }}
            animate={{
              d: data
                .map(
                  (d, i) =>
                    `${i === 0 ? 'M' : 'L'} ${i * (barWidth + 6) + barWidth / 2},${chartHeight - (d.value / maxVal) * chartHeight * 0.85}`,
                )
                .join(' '),
            }}
            transition={{ duration: 0.5, ease: 'easeOut' }}
            fill="none"
            stroke="hsl(168, 76%, 42%)"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />

          {/* Dots */}
          {data.map((d, i) => (
            <motion.circle
              key={i}
              initial={{ r: 0 }}
              animate={{ r: 3 }}
              transition={{ delay: 0.4 + i * 0.03, duration: 0.2 }}
              cx={i * (barWidth + 6) + barWidth / 2}
              cy={chartHeight - (d.value / maxVal) * chartHeight * 0.85}
              fill="hsl(168, 76%, 42%)"
              className="stroke-white"
              strokeWidth="2"
            />
          ))}
        </svg>
      </div>

      {/* Labels */}
      <div
        className="flex mt-2"
        style={{ gap: '6px' }}
      >
        {data.map((d, i) => (
          <span
            key={i}
            className="text-[10px] text-muted-foreground text-center truncate"
            style={{ width: barWidth }}
          >
            {d.label}
          </span>
        ))}
      </div>
    </motion.div>
  )
}
