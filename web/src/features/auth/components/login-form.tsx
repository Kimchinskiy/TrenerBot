'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from 'next/navigation'
import { Button, Input, Label, Card } from '@/components/ui'
import { ScreenHeader, ErrorBox } from '@/components/ui/screen'
import { loginSchema, type LoginValues } from '../schemas'
import { loginWithPassword, loginWithTelegramWidget } from '@/lib/auth'
import { useAuth } from '@/components/auth-provider'
import { TelegramLoginButton } from '@/services/telegram'
import { BOT_USERNAME } from './config'

export function LoginForm() {
  const router = useRouter()
  const { setAuthed } = useAuth()
  const [serverError, setServerError] = useState('')
  const [busy, setBusy] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginValues>({ resolver: zodResolver(loginSchema) })

  const onSubmit = handleSubmit(async (values) => {
    setServerError('')
    setBusy(true)
    try {
      await loginWithPassword(values.phone.trim(), values.password)
      setAuthed()
      router.replace('/dashboard')
    } catch (e) {
      setServerError(e instanceof Error ? e.message : 'Ошибка входа')
    } finally {
      setBusy(false)
    }
  })

  const onTelegram = async (fields: Record<string, string>) => {
    setServerError('')
    setBusy(true)
    try {
      await loginWithTelegramWidget(fields)
      setAuthed()
      router.replace('/dashboard')
    } catch (e) {
      setServerError(e instanceof Error ? e.message : 'Ошибка входа через Telegram')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col justify-center px-4 py-8">
      <div className="w-full max-w-md mx-auto">
        <ScreenHeader title="Вход" />
        
        <Card className="mt-2 flex flex-col gap-4 shadow-lg border-border/80">
          {serverError && <ErrorBox error={new Error(serverError)} />}
          
          <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <div>
              <Label htmlFor="phone">Телефон</Label>
              <Input
                id="phone"
                placeholder="+7 (999) 123-45-67"
                inputMode="tel"
                autoComplete="tel"
                {...register('phone')}
              />
              {errors.phone && (
                <p className="mt-1.5 text-xs font-semibold text-destructive">{errors.phone.message}</p>
              )}
            </div>
            
            <div>
              <Label htmlFor="password">Пароль</Label>
              <Input
                id="password"
                type="password"
                placeholder="Введите пароль"
                autoComplete="current-password"
                {...register('password')}
              />
              {errors.password && (
                <p className="mt-1.5 text-xs font-semibold text-destructive">{errors.password.message}</p>
              )}
            </div>
            
            <Button type="submit" disabled={busy} className="mt-2 font-bold">
              {busy ? 'Подождите...' : 'Войти'}
            </Button>
          </form>

          <button
            type="button"
            onClick={() => router.push('/register')}
            className="py-2 text-center text-sm font-semibold text-primary hover:underline hover:opacity-90 active:scale-95 transition-all"
          >
            Нет аккаунта? Зарегистрироваться
          </button>

          {BOT_USERNAME && (
            <>
              <div className="flex items-center gap-3 py-1 text-muted-foreground">
                <div className="h-px flex-1 bg-border" />
                <span className="text-xs font-medium uppercase tracking-wider">или</span>
                <div className="h-px flex-1 bg-border" />
              </div>
              <div className="flex justify-center pt-1">
                <TelegramLoginButton onAuth={onTelegram} />
              </div>
            </>
          )}
        </Card>
      </div>
    </div>
  )
}
