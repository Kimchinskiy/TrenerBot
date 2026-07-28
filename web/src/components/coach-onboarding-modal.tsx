'use client'

import { useState, useEffect } from 'react'
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
  open: boolean
  onClose: () => void
  userId?: number
}

export default function CoachOnboardingModal({ open, onClose, userId }: Props) {
  const [step, setStep] = useState(1)

  useEffect(() => {
    if (open) {
      setStep(1)
    }
  }, [open])

  if (!open) return null

  const handleNext = () => {
    if (step < 4) {
      setStep(step + 1)
    } else {
      if (userId) {
        localStorage.setItem(`plavli_coach_onboarding_done_${userId}`, 'true')
      }
      localStorage.setItem('plavli_coach_onboarding_done', 'true')
      onClose()
    }
  }

  const handlePrev = () => {
    if (step > 1) {
      setStep(step - 1)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-md p-4 animate-in fade-in duration-200">
      <div className="w-full max-w-lg rounded-3xl bg-card border border-border/60 p-6 shadow-2xl relative overflow-hidden flex flex-col gap-6">
        {/* Top Progress bar */}
        <div className="flex items-center justify-between gap-3">
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
            onClick={onClose}
            className="rounded-xl p-1.5 hover:bg-muted/50 text-muted-foreground transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Card Content based on step */}
        {step === 1 && (
          <div className="flex flex-col gap-4 animate-in fade-in duration-300">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
              <Sparkles className="h-7 w-7" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-foreground mb-2">
                Добро пожаловать в Плавли 👋
              </h2>
              <p className="text-sm font-semibold text-primary mb-3">
                Это единое рабочее пространство тренера.
              </p>
              <p className="text-sm text-muted-foreground leading-relaxed mb-3">
                Больше не нужно искать информацию в чатах, блокнотах и таблицах — всё, что связано с тренировками и клиентами, собрано в одном месте.
              </p>
              <p className="text-sm text-muted-foreground leading-relaxed">
                В несколько касаний вы сможете управлять всей своей группой.
              </p>
            </div>
            <Button onClick={handleNext} variant="gradient" size="lg" className="w-full mt-2 rounded-2xl">
              Начать <ChevronRight className="h-4 w-4 ml-1" />
            </Button>
          </div>
        )}

        {step === 2 && (
          <div className="flex flex-col gap-4 animate-in fade-in duration-300">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-success-light text-success">
              <CheckCircle2 className="h-7 w-7" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-foreground mb-1">
                Почему это удобно?
              </h2>
              <p className="text-xs font-semibold text-muted-foreground mb-4">
                Меньше рутины — больше времени на тренировки
              </p>
              <p className="text-xs text-muted-foreground mb-3">
                Плавли помогает избавиться от ежедневной рутины.
              </p>

              <div className="grid grid-cols-1 gap-2.5 my-2">
                {[
                  'Расписание всегда под рукой',
                  'Быстрая отметка посещаемости',
                  'Карточки клиентов со всей информацией',
                  'Автоматические напоминания клиентам',
                  'Массовые уведомления всей группе',
                  'История тренировок и посещений',
                ].map((item, idx) => (
                  <div key={idx} className="flex items-center gap-2.5 text-sm text-foreground">
                    <div className="flex h-5 w-5 items-center justify-center rounded-full bg-success-light text-success shrink-0">
                      <Check className="h-3 w-3" />
                    </div>
                    <span>{item}</span>
                  </div>
                ))}
              </div>

              <div className="mt-4 p-3 rounded-2xl bg-muted/40 text-center text-xs font-semibold text-muted-foreground border border-border/40">
                Никаких таблиц, переписок и забытых сообщений.
              </div>
            </div>
            <div className="flex gap-2 mt-2">
              <Button onClick={handlePrev} variant="outline" size="lg" className="flex-1 rounded-2xl">
                <ChevronLeft className="h-4 w-4 mr-1" /> Назад
              </Button>
              <Button onClick={handleNext} variant="gradient" size="lg" className="flex-1 rounded-2xl">
                Далее <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        )}

        {step === 3 && (
          <div className="flex flex-col gap-4 animate-in fade-in duration-300">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-cyan-500/10 text-cyan-500">
              <Send className="h-7 w-7" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-foreground mb-1">
                Telegram-бот
              </h2>
              <p className="text-xs font-semibold text-primary mb-3">
                Веб-приложение + Telegram = идеальная связка
              </p>
              <p className="text-xs text-muted-foreground leading-relaxed mb-4">
                Пока вы работаете в приложении, бот помогает автоматически взаимодействовать с клиентами.
              </p>

              <p className="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-2">
                Он может:
              </p>
              <div className="grid grid-cols-1 gap-2.5 mb-4">
                <div className="flex items-center gap-3 p-2.5 rounded-2xl bg-card border border-border/50 text-sm">
                  <span className="text-lg">📲</span>
                  <span className="text-foreground">напоминать о тренировках</span>
                </div>
                <div className="flex items-center gap-3 p-2.5 rounded-2xl bg-card border border-border/50 text-sm">
                  <span className="text-lg">📢</span>
                  <span className="text-foreground">отправлять ваши объявления</span>
                </div>
                <div className="flex items-center gap-3 p-2.5 rounded-2xl bg-card border border-border/50 text-sm">
                  <span className="text-lg">📅</span>
                  <span className="text-foreground">показывать расписание</span>
                </div>
                <div className="flex items-center gap-3 p-2.5 rounded-2xl bg-card border border-border/50 text-sm">
                  <span className="text-lg">🔔</span>
                  <span className="text-foreground">уведомлять об изменениях</span>
                </div>
              </div>

              <p className="text-xs text-center font-medium text-muted-foreground">
                Вам не нужно писать каждому человеку вручную.
              </p>
            </div>
            <div className="flex gap-2 mt-2">
              <Button onClick={handlePrev} variant="outline" size="lg" className="flex-1 rounded-2xl">
                <ChevronLeft className="h-4 w-4 mr-1" /> Назад
              </Button>
              <Button onClick={handleNext} variant="gradient" size="lg" className="flex-1 rounded-2xl">
                Далее <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          </div>
        )}

        {step === 4 && (
          <div className="flex flex-col gap-4 animate-in fade-in duration-300">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
              <Rocket className="h-7 w-7" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-foreground mb-2">
                Всё готово! 🚀
              </h2>
              <p className="text-sm text-foreground leading-relaxed mb-3 font-semibold">
                Теперь можно начать пользоваться системой.
              </p>
              <p className="text-sm text-muted-foreground leading-relaxed">
                Добавьте первых клиентов, настройте расписание и сосредоточьтесь на тренировках — о рутине позаботится Плавли.
              </p>
            </div>
            <div className="flex gap-2 mt-4">
              <Button onClick={handlePrev} variant="outline" size="lg" className="rounded-2xl px-4">
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <Button onClick={handleNext} variant="gradient" size="lg" className="flex-1 rounded-2xl">
                Перейти в приложение <Rocket className="h-4 w-4 ml-2" />
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
