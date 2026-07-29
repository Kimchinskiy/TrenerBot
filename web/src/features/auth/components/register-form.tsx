'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from 'next/navigation'
import { Button, Input, Label, Card } from '@/components/ui'
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
    <div className="min-h-screen flex flex-col justify-center px-4 py-8">
      <div className="w-full max-w-md mx-auto">
        <ScreenHeader title="Регистрация" subtitle="Создайте личный кабинет CRM" />
        
        <Card className="mt-2 flex flex-col gap-4 shadow-lg border-border/80">
          {serverError && <ErrorBox error={new Error(serverError)} />}
          
          <form onSubmit={onSubmit} className="flex flex-col gap-4">
            <div>
              <Label htmlFor="firstName">Имя</Label>
              <Input
                id="firstName"
                placeholder="Климент"
                autoComplete="given-name"
                {...register('firstName')}
              />
              {errors.firstName && (
                <p className="mt-1.5 text-xs font-semibold text-destructive">{errors.firstName.message}</p>
              )}
            </div>
            
            <div>
              <Label htmlFor="lastName">Фамилия</Label>
              <Input
                id="lastName"
                placeholder="Ворошилов"
                autoComplete="family-name"
                {...register('lastName')}
              />
              {errors.lastName && (
                <p className="mt-1.5 text-xs font-semibold text-destructive">{errors.lastName.message}</p>
              )}
            </div>
            
            <div>
              <Label htmlFor="phone">Телефон</Label>
              <Input
                id="phone"
    placeholder="+7 (999) 123-45-67"
    inputMode="tel"
    autoComplete="tel"
    maxLength={12}
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
                placeholder="Минимум 6 символов"
                autoComplete="new-password"
                {...register('password')}
              />
              {errors.password && (
                <p className="mt-1.5 text-xs font-semibold text-destructive">{errors.password.message}</p>
              )}
            </div>
            
            <Button type="submit" disabled={busy} className="mt-2 font-bold">
              {busy ? 'Подождите...' : 'Зарегистрироваться'}
            </Button>
          </form>

          <button
            type="button"
            onClick={() => router.push('/login')}
            className="py-2 text-center text-sm font-semibold text-primary hover:underline hover:opacity-90 active:scale-95 transition-all"
          >
            Уже есть аккаунт? Войти
          </button>
        </Card>
      </div>
    </div>
  )
}
