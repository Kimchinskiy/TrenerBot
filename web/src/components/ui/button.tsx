import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center whitespace-nowrap rounded-2xl text-base font-semibold transition-all duration-200 active:scale-[0.97] disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        primary:
          'bg-primary text-primary-foreground hover:bg-primary/90 shadow-md shadow-primary/20',
        secondary:
          'bg-secondary text-secondary-foreground hover:bg-secondary/80 border border-border/50',
        destructive:
          'bg-destructive text-destructive-foreground hover:bg-destructive/90 shadow-sm',
        outline:
          'border border-border bg-white text-foreground hover:bg-muted/50',
        ghost:
          'bg-transparent text-foreground hover:bg-muted/50',
        gradient:
          'bg-gradient-to-r from-primary to-cyan-500 text-white hover:opacity-90 shadow-md shadow-primary/25',
      },
      size: {
        default: 'w-full px-4 py-3 h-12',
        sm: 'px-3 py-1.5 text-sm h-9 rounded-xl',
        lg: 'w-full px-6 py-4 h-14 text-lg rounded-2xl',
        icon: 'h-10 w-10 rounded-xl',
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
