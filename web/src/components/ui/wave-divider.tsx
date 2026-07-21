import { cn } from '@/lib/utils'

interface WaveDividerProps {
  className?: string
  flip?: boolean
}

export function WaveDivider({ className, flip }: WaveDividerProps) {
  return (
    <div className={cn('w-full overflow-hidden leading-none', flip && 'rotate-180', className)}>
      <svg
        viewBox="0 0 1200 60"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className="w-full h-auto"
        preserveAspectRatio="none"
      >
        <path
          d="M0 30C200 10 400 50 600 30C800 10 1000 50 1200 30V60H0V30Z"
          fill="currentColor"
          className="text-primary/5"
        />
      </svg>
    </div>
  )
}
