'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from 'next/navigation'
import { Button, Input, Label } from '@/components/ui'
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
    <div>
      <ScreenHeader title="Вход" subtitle="Спортивная CRM" />
      <div className="flex flex-col gap-3 px-4 pb-8">
        {serverError && (
          <div className="rounded-xl bg-red-500/15 p-3 text-sm text-red-300">{serverError}</div>
        )}
        <form onSubmit={onSubmit} className="flex flex-col gap-3">
          <div>
            <Label htmlFor="phone">Телефон</Label>
            <Input id="phone" placeholder="Телефон, например +79991234567" inputMode="tel" autoComplete="tel" {...register('phone')} />
            {errors.phone && <p className="mt-1 text-xs text-red-400">{errors.phone.message}</p>}
          </div>
          <div>
            <Label htmlFor="password">Пароль</Label>
            <Input id="password" type="password" placeholder="Пароль" autoComplete="current-password" {...register('password')} />
            {errors.password && <p className="mt-1 text-xs text-red-400">{errors.password.message}</p>}
          </div>
          <Button type="submit" disabled={busy}>
            {busy ? 'Подождите...' : 'Войти'}
          </Button>
        </form>

        <button
          type="button"
          onClick={() => router.push('/register')}
          className="py-1 text-center text-sm text-tg-link"
        >
          Нет аккаунта? Зарегистрироваться
        </button>

        {BOT_USERNAME && (
          <>
            <div className="flex items-center gap-3 py-1 text-tg-hint">
              <div className="h-px flex-1 bg-tg-secondary" />
              <span className="text-xs">или</span>
              <div className="h-px flex-1 bg-tg-secondary" />
            </div>
            <TelegramLoginButton onAuth={onTelegram} />
          </>
        )}
      </div>
    </div>
  )
}
