'use client'

import * as React from 'react'
import { Accordion as BaseAccordion } from '@base-ui/react/accordion'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'

const AccordionRoot = React.forwardRef<
  HTMLDivElement,
  BaseAccordion.Root.Props<unknown>
>(({ className, ...props }, ref) => (
  <BaseAccordion.Root
    ref={ref}
    className={cn('w-full', className)}
    {...props}
  />
))
AccordionRoot.displayName = 'Accordion'

const AccordionItem = React.forwardRef<
  HTMLDivElement,
  BaseAccordion.Item.Props
>(({ className, ...props }, ref) => (
  <BaseAccordion.Item
    ref={ref}
    className={cn('border-b border-border last:border-b-0', className)}
    {...props}
  />
))
AccordionItem.displayName = 'AccordionItem'

const AccordionTrigger = React.forwardRef<
  HTMLButtonElement,
  BaseAccordion.Trigger.Props
>(({ className, children, ...props }, ref) => (
  <BaseAccordion.Header className="flex">
    <BaseAccordion.Trigger
      ref={ref}
      className={cn(
        'flex flex-1 items-center justify-between py-4 text-sm font-bold text-left transition-all hover:opacity-80 [&[data-panel-open]>svg]:rotate-180',
        className,
      )}
      {...props}
    >
      {children}
      <ChevronDown className="h-4 w-4 shrink-0 transition-transform duration-200 text-muted-foreground" />
    </BaseAccordion.Trigger>
  </BaseAccordion.Header>
))
AccordionTrigger.displayName = 'AccordionTrigger'

const AccordionPanel = React.forwardRef<
  HTMLDivElement,
  BaseAccordion.Panel.Props
>(({ className, children, ...props }, ref) => (
  <BaseAccordion.Panel
    ref={ref}
    className={cn(
      'overflow-hidden text-sm text-muted-foreground transition-all data-[panel-open]:animate-in data-[panel-open]:slide-in-from-top-1 data-[panel-open]:fade-in-0 data-[panel-open]:duration-200',
      className,
    )}
    {...props}
  >
    <div className="pb-4 pt-0 whitespace-pre-wrap leading-relaxed">{children}</div>
  </BaseAccordion.Panel>
))
AccordionPanel.displayName = 'AccordionPanel'

export const Accordion = Object.assign(AccordionRoot, {
  Item: AccordionItem,
  Trigger: AccordionTrigger,
  Panel: AccordionPanel,
})
