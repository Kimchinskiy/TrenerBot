'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from 'next/navigation'
import { Button, Input, Label } from '@/components/ui'
import { ScreenHeader, ErrorBox } from '@/components/ui/screen'
import { registerSchema, type RegisterValues } from '../schemas'
import { registerWithPassword } from '@/lib/auth'
import { useAuth } from '@/components/auth-provider'

export function RegisterForm() {
  const router = useRouter()
  const { setAuthed } = useAuth()
  const [serverError, setServerError] = useState('')
  const [busy, setBusy] = useState(false)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterValues>({ resolver: zodResolver(registerSchema) })

  const onSubmit = handleSubmit(async (values) => {
    setServerError('')
    setBusy(true)
    try {
      await registerWithPassword(
        values.phone.trim(),
        values.password,
        values.firstName.trim(),
        (values.lastName ?? '').trim(),
      )
      setAuthed()
      router.replace('/dashboard')
    } catch (e) {
      setServerError(e instanceof Error ? e.message : 'Ошибка регистрации')
    } finally {
      setBusy(false)
    }
  })

  return (
    <div>
      <ScreenHeader title="Регистрация" subtitle="Спортивная CRM" />
      <div className="flex flex-col gap-3 px-4 pb-8">
        {serverError && (
          <div className="rounded-xl bg-red-500/15 p-3 text-sm text-red-300">{serverError}</div>
        )}
        <form onSubmit={onSubmit} className="flex flex-col gap-3">
          <div>
            <Label htmlFor="firstName">Имя</Label>
            <Input id="firstName" placeholder="Имя" autoComplete="given-name" {...register('firstName')} />
            {errors.firstName && <p className="mt-1 text-xs text-red-400">{errors.firstName.message}</p>}
          </div>
          <div>
            <Label htmlFor="lastName">Фамилия</Label>
            <Input id="lastName" placeholder="Фамилия" autoComplete="family-name" {...register('lastName')} />
            {errors.lastName && <p className="mt-1 text-xs text-red-400">{errors.lastName.message}</p>}
          </div>
          <div>
            <Label htmlFor="phone">Телефон</Label>
            <Input id="phone" placeholder="Телефон, например +79991234567" inputMode="tel" autoComplete="tel" {...register('phone')} />
            {errors.phone && <p className="mt-1 text-xs text-red-400">{errors.phone.message}</p>}
          </div>
          <div>
            <Label htmlFor="password">Пароль</Label>
            <Input id="password" type="password" placeholder="Пароль (минимум 6 символов)" autoComplete="new-password" {...register('password')} />
            {errors.password && <p className="mt-1 text-xs text-red-400">{errors.password.message}</p>}
          </div>
          <Button type="submit" disabled={busy}>
            {busy ? 'Подождите...' : 'Зарегистрироваться'}
          </Button>
        </form>

        <button
          type="button"
          onClick={() => router.push('/login')}
          className="py-1 text-center text-sm text-tg-link"
        >
          Уже есть аккаунт? Войти
        </button>
      </div>
    </div>
  )
}
