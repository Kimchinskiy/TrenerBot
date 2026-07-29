'use client'

import { useState } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Send, CheckCircle2, AlertCircle, ExternalLink, Copy, Check, X } from 'lucide-react'

interface Props {
  userId?: number
  hasBotAccess?: boolean
  phone?: string
}

export function TelegramIntegrationCard({ userId, hasBotAccess, phone }: Props) {
  const [showModal, setShowModal] = useState(false)
  const [copied, setCopied] = useState(false)

  const botUsername = 'plavli_bot'
  const bindCode = userId ? `bind_${userId}` : 'bind'
  const telegramUrl = `https://t.me/${botUsername}?start=${bindCode}`

  const handleCopy = () => {
    navigator.clipboard.writeText(telegramUrl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <>
      <Card className="p-4 border-border/50 bg-card relative overflow-hidden flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-cyan-500/10 text-cyan-500 shrink-0">
              <Send className="h-5 w-5" />
            </div>
            <div>
              <h3 className="text-sm font-bold text-foreground">Telegram-бот</h3>
              <p className="text-xs text-muted-foreground">@{botUsername}</p>
            </div>
          </div>

          <span
            className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-bold ${
              hasBotAccess
                ? 'bg-success-light text-success'
                : 'bg-muted text-muted-foreground'
            }`}
          >
            {hasBotAccess ? (
              <>
                <CheckCircle2 className="h-3.5 w-3.5" />
                <span>Подключён</span>
              </>
            ) : (
              <>
                <AlertCircle className="h-3.5 w-3.5" />
                <span>Не подключён</span>
              </>
            )}
          </span>
        </div>

        <p className="text-xs text-muted-foreground leading-relaxed">
          {hasBotAccess
            ? 'Ваш аккаунт синхронизирован с Telegram-ботом. Вы будете получать уведомления и напоминания о тренировках.'
            : 'Подключите официального бота @plavli_bot для мгновенного получения уведомлений и связи с клиентами.'}
        </p>

        {hasBotAccess ? (
          <a
            href={`https://t.me/${botUsername}`}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center justify-center gap-2 rounded-2xl bg-primary/10 hover:bg-primary/20 text-primary py-2.5 px-4 text-xs font-bold transition-all"
          >
            <span>Открыть @{botUsername}</span>
            <ExternalLink className="h-3.5 w-3.5" />
          </a>
        ) : (
          <Button
            onClick={() => setShowModal(true)}
            variant="gradient"
            size="sm"
            className="rounded-2xl w-full"
          >
            <Send className="h-3.5 w-3.5 mr-2" />
            Подключить Telegram
          </Button>
        )}
      </Card>

      {/* Connect Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-200">
          <div className="w-full max-w-md rounded-3xl bg-card border border-border/60 p-6 shadow-2xl relative flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-cyan-500/10 text-cyan-500">
                  <Send className="h-5 w-5" />
                </div>
                <h2 className="text-base font-bold text-foreground">Подключение @{botUsername}</h2>
              </div>
              <button
                onClick={() => setShowModal(false)}
                className="rounded-xl p-1.5 hover:bg-muted/50 text-muted-foreground transition-colors"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="flex flex-col gap-3 my-1">
              <p className="text-xs text-muted-foreground">
                Выполните 2 простых шага для привязки аккаунта:
              </p>

              <div className="flex items-start gap-3 p-3 rounded-2xl bg-muted/40 border border-border/40 text-xs">
                <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary text-white text-[11px] font-bold shrink-0">
                  1
                </span>
                <div className="text-foreground">
                  Нажмите на кнопку ниже или откройте ссылку: <span className="font-semibold text-primary">@{botUsername}</span>
                </div>
              </div>

              <div className="flex items-start gap-3 p-3 rounded-2xl bg-muted/40 border border-border/40 text-xs">
                <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary text-white text-[11px] font-bold shrink-0">
                  2
                </span>
                <div className="text-foreground">
                  В открывшемся диалоге нажмите кнопку <span className="font-bold text-primary">Запустить</span> (или отправьте кодовую команду).
                </div>
              </div>
            </div>

            {/* Copyable Link */}
            <div className="flex items-center justify-between gap-2 p-2.5 rounded-2xl bg-card border border-border/60 text-xs">
              <span className="truncate font-mono text-muted-foreground px-1">{telegramUrl}</span>
              <button
                onClick={handleCopy}
                className="flex items-center gap-1 px-2.5 py-1.5 rounded-xl bg-muted hover:bg-muted/80 text-foreground font-semibold shrink-0 transition-colors"
              >
                {copied ? <Check className="h-3.5 w-3.5 text-success" /> : <Copy className="h-3.5 w-3.5" />}
                <span>{copied ? 'Скопировано' : 'Копировать'}</span>
              </button>
            </div>

            <a
              href={telegramUrl}
              target="_blank"
              rel="noopener noreferrer"
              onClick={() => setShowModal(false)}
              className="flex items-center justify-center gap-2 rounded-2xl bg-gradient-to-r from-primary to-cyan-500 text-white py-3 px-4 text-sm font-bold shadow-md hover:opacity-90 active:scale-95 transition-all mt-1"
            >
              <Send className="h-4 w-4" />
              <span>Открыть Telegram-бота</span>
            </a>
          </div>
        </div>
      )}
    </>
  )
}
