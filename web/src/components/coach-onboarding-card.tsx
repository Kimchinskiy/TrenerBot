'use client'

import { useState } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Sparkles,
  CheckCircle2,
  Send,
  Rocket,
  ChevronRight,
  ChevronLeft,
  X,
  Check,
} from 'lucide-react'

interface Props {
  onFinish?: () => void
  userId?: number
}

export function CoachOnboardingCard({ onFinish, userId }: Props) {
  const [step, setStep] = useState(1)
  const [dismissed, setDismissed] = useState(false)

  if (dismissed) return null

  const handleNext = () => {
    if (step < 4) {
      setStep(step + 1)
    } else {
      if (userId) {
        localStorage.setItem(`plavli_coach_onboarding_done_${userId}`, 'true')
      }
      localStorage.setItem('plavli_coach_onboarding_done', 'true')
      setDismissed(true)
      if (onFinish) onFinish()
    }
  }

  const handlePrev = () => {
    if (step > 1) {
      setStep(step - 1)
    }
  }

  const handleClose = () => {
    if (userId) {
      localStorage.setItem(`plavli_coach_onboarding_done_${userId}`, 'true')
    }
    localStorage.setItem('plavli_coach_onboarding_done', 'true')
    setDismissed(true)
    if (onFinish) onFinish()
  }

  return (
    <Card className="relative overflow-hidden border-primary/30 bg-gradient-to-br from-card via-card to-primary/5 p-5 shadow-elevated animate-in fade-in duration-300">
      {/* Top progress indicators */}
      <div className="flex items-center justify-between gap-3 mb-4">
        <div className="flex items-center gap-1.5 flex-1">
          {[1, 2, 3, 4].map((s) => (
            <div
              key={s}
              className={`h-1.5 flex-1 rounded-full transition-all duration-300 ${
                s <= step ? 'bg-primary' : 'bg-muted/60'
              }`}
            />
          ))}
        </div>
        <button
          onClick={handleClose}
          className="rounded-xl p-1 hover:bg-muted/50 text-muted-foreground transition-colors"
          title="Скрыть"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      {/* Card 1: Welcome */}
      {step === 1 && (
        <div className="flex flex-col gap-3 animate-in fade-in duration-300">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 text-primary shrink-0">
              <Sparkles className="h-6 w-6" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-foreground">
                Добро пожаловать в Плавли 👋
              </h3>
              <p className="text-xs font-semibold text-primary">
                Это единое рабочее пространство тренера.
              </p>
            </div>
          </div>

          <p className="text-sm text-muted-foreground leading-relaxed">
            Больше не нужно искать информацию в чатах, блокнотах и таблицах — всё, что связано с тренировками и клиентами, собрано в одном месте.
          </p>

          <p className="text-xs font-medium text-muted-foreground">
            В несколько касаний вы сможете управлять всей своей группой.
          </p>

          <Button onClick={handleNext} variant="gradient" size="lg" className="w-full mt-2 rounded-2xl">
            Начать <ChevronRight className="h-4 w-4 ml-1" />
          </Button>
        </div>
      )}

      {/* Card 2: Benefits */}
      {step === 2 && (
        <div className="flex flex-col gap-3 animate-in fade-in duration-300">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-success-light text-success shrink-0">
              <CheckCircle2 className="h-6 w-6" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-foreground">
                Почему это удобно?
              </h3>
              <p className="text-xs font-semibold text-muted-foreground">
                Меньше рутины — больше времени на тренировки
              </p>
            </div>
          </div>

          <p className="text-xs text-muted-foreground">
            Плавли помогает избавиться от ежедневной рутины.
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 my-1">
            {[
              'Расписание всегда под рукой',
              'Быстрая отметка посещаемости',
              'Карточки клиентов со всей информацией',
              'Автоматические напоминания клиентам',
              'Массовые уведомления всей группе',
              'История тренировок и посещений',
            ].map((item, idx) => (
              <div key={idx} className="flex items-center gap-2 text-xs font-medium text-foreground">
                <div className="flex h-4 w-4 items-center justify-center rounded-full bg-success-light text-success shrink-0">
                  <Check className="h-2.5 w-2.5" />
                </div>
                <span>{item}</span>
              </div>
            ))}
          </div>

          <div className="p-2.5 rounded-2xl bg-muted/40 text-center text-xs font-semibold text-muted-foreground border border-border/40">
            Никаких таблиц, переписок и забытых сообщений.
          </div>

          <div className="flex gap-2 mt-1">
            <Button onClick={handlePrev} variant="outline" size="sm" className="flex-1 rounded-2xl">
              <ChevronLeft className="h-4 w-4 mr-1" /> Назад
            </Button>
            <Button onClick={handleNext} variant="gradient" size="sm" className="flex-1 rounded-2xl">
              Далее <ChevronRight className="h-4 w-4 ml-1" />
            </Button>
          </div>
        </div>
      )}

      {/* Card 3: Telegram Bot */}
      {step === 3 && (
        <div className="flex flex-col gap-3 animate-in fade-in duration-300">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-cyan-500/10 text-cyan-500 shrink-0">
              <Send className="h-6 w-6" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-foreground">
                Telegram-бот
              </h3>
              <p className="text-xs font-semibold text-primary">
                Веб-приложение + Telegram = идеальная связка
              </p>
            </div>
          </div>

          <p className="text-xs text-muted-foreground leading-relaxed">
            Пока вы работаете в приложении, бот помогает автоматически взаимодействовать с клиентами.
          </p>

          <p className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
            Он может:
          </p>

          <div className="grid grid-cols-2 gap-2">
            <div className="flex items-center gap-2 p-2 rounded-2xl bg-card border border-border/50 text-xs">
              <span>📲</span>
              <span className="text-foreground font-medium">напоминать о тренировках</span>
            </div>
            <div className="flex items-center gap-2 p-2 rounded-2xl bg-card border border-border/50 text-xs">
              <span>📢</span>
              <span className="text-foreground font-medium">отправлять объявления</span>
            </div>
            <div className="flex items-center gap-2 p-2 rounded-2xl bg-card border border-border/50 text-xs">
              <span>📅</span>
              <span className="text-foreground font-medium">показывать расписание</span>
            </div>
            <div className="flex items-center gap-2 p-2 rounded-2xl bg-card border border-border/50 text-xs">
              <span>🔔</span>
              <span className="text-foreground font-medium">уведомлять об изменениях</span>
            </div>
          </div>

          <p className="text-xs text-center font-medium text-muted-foreground">
            Вам не нужно писать каждому человеку вручную.
          </p>

          <div className="flex gap-2 mt-1">
            <Button onClick={handlePrev} variant="outline" size="sm" className="flex-1 rounded-2xl">
              <ChevronLeft className="h-4 w-4 mr-1" /> Назад
            </Button>
            <Button onClick={handleNext} variant="gradient" size="sm" className="flex-1 rounded-2xl">
              Далее <ChevronRight className="h-4 w-4 ml-1" />
            </Button>
          </div>
        </div>
      )}

      {/* Card 4: Ready */}
      {step === 4 && (
        <div className="flex flex-col gap-3 animate-in fade-in duration-300">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 text-primary shrink-0">
              <Rocket className="h-6 w-6" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-foreground">
                Всё готово! 🚀
              </h3>
              <p className="text-xs font-semibold text-foreground">
                Теперь можно начать пользоваться системой.
              </p>
            </div>
          </div>

          <p className="text-sm text-muted-foreground leading-relaxed">
            Добавьте первых клиентов, настройте расписание и сосредоточьтесь на тренировках — о рутине позаботится Плавли.
          </p>

          <div className="flex gap-2 mt-2">
            <Button onClick={handlePrev} variant="outline" size="sm" className="rounded-2xl px-3">
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button onClick={handleNext} variant="gradient" size="sm" className="flex-1 rounded-2xl">
              Перейти в приложение <Rocket className="h-4 w-4 ml-1.5" />
            </Button>
          </div>
        </div>
      )}
    </Card>
  )
}
