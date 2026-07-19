import * as React from 'react'
import { cn } from '@/lib/utils'

const Card = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & { onClick?: () => void }>(
  ({ className, onClick, ...props }, ref) => (
    <div
      ref={ref}
      onClick={onClick}
      className={cn(
        'rounded-2xl border border-border bg-card p-5 text-card-foreground shadow-sm transition-all duration-200',
        onClick ? 'cursor-pointer hover:border-foreground/20 active:scale-[0.99]' : '',
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
