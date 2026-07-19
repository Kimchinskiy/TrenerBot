import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center whitespace-nowrap rounded-xl text-base font-semibold transition-all duration-200 active:scale-[0.98] disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        primary: 'bg-primary text-primary-foreground hover:opacity-95 shadow-sm',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/90 border border-border/40',
        destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/95 shadow-sm',
        outline: 'border border-border bg-transparent text-foreground hover:bg-muted/30',
        ghost: 'bg-transparent text-foreground hover:bg-muted/30',
      },
      size: {
        default: 'w-full px-4 py-3 h-12',
        sm: 'px-3 py-1.5 text-sm h-9',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'default',
    },
  },
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button'
    return (
      <Comp className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />
    )
  },
)
Button.displayName = 'Button'

export { Button, buttonVariants }
