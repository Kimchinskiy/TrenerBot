import * as React from 'react'
import { cn } from '@/lib/utils'

const Card = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & { onClick?: () => void; glass?: boolean }>(
  ({ className, onClick, glass, ...props }, ref) => (
    <div
      ref={ref}
      onClick={onClick}
      className={cn(
        'rounded-3xl border bg-white p-5 text-card-foreground transition-all duration-200',
        glass
          ? 'glass-card shadow-card border-white/60'
          : 'shadow-card border-border/50',
        onClick ? 'cursor-pointer hover:shadow-elevated active:scale-[0.99]' : '',
        className,
      )}
      {...props}
    />
  ),
)
Card.displayName = 'Card'

const CardHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn('flex flex-col space-y-1.5 pb-3', className)} {...props} />
  ),
)
CardHeader.displayName = 'CardHeader'

const CardContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => <div ref={ref} className={cn('pt-0', className)} {...props} />,
)
CardContent.displayName = 'CardContent'

export { Card, CardHeader, CardContent }
