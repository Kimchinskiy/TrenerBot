import { useEffect, useRef, useState } from 'react'
import { Button, ScreenHeader } from '../components/ui'
import {
  loginWithPassword,
  registerWithPassword,
  loginWithTelegramWidget,
} from '../lib/auth'

const BOT_USERNAME = import.meta.env.VITE_TELEGRAM_BOT || ''

type Mode = 'login' | 'register'

// TelegramLoginButton injects the official Telegram Login Widget and forwards the
// signed user payload to the backend. Rendered only when a bot username is set.
function TelegramLoginButton({ onAuth }: { onAuth: (fields: Record<string, string>) => void }) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!BOT_USERNAME || !ref.current) return
    ;(window as any).onTelegramAuth = (user: Record<string, unknown>) => {
      const fields: Record<string, string> = {}
      Object.entries(user).forEach(([k, v]) => (fields[k] = String(v)))
      onAuth(fields)
    }
    const script = document.createElement('script')
    script.src = 'https://telegram.org/js/telegram-widget.js?22'
    script.async = true
    script.setAttribute('data-telegram-login', BOT_USERNAME)
    script.setAttribute('data-size', 'large')
    script.setAttribute('data-radius', '12')
    script.setAttribute('data-onauth', 'onTelegramAuth(user)')
    script.setAttribute('data-request-access', 'write')
    ref.current.appendChild(script)
    return () => {
      if (ref.current) ref.current.innerHTML = ''
    }
  }, [onAuth])

  if (!BOT_USERNAME) return null
  return <div ref={ref} className="flex justify-center pt-2" />
}

export default function Login({ onSuccess }: { onSuccess: () => void }) {
  const [mode, setMode] = useState<Mode>('login')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const submit = async () => {
    setError('')
    if (!phone.trim() || !password) {
      setError('Введите телефон и пароль')
      return
    }
    setBusy(true)
    try {
      if (mode === 'login') {
        await loginWithPassword(phone.trim(), password)
      } else {
        await registerWithPassword(phone.trim(), password, firstName.trim(), lastName.trim())
      }
      onSuccess()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка входа')
    } finally {
      setBusy(false)
    }
  }

  const onTelegram = async (fields: Record<string, string>) => {
    setError('')
    setBusy(true)
    try {
      await loginWithTelegramWidget(fields)
      onSuccess()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка входа через Telegram')
    } finally {
      setBusy(false)
    }
  }

  const inputCls = 'w-full rounded-xl bg-tg-secondary p-3 text-tg-text outline-none placeholder:text-tg-hint'

  return (
    <div className="mx-auto flex min-h-full max-w-md flex-col">
      <ScreenHeader
        title={mode === 'login' ? 'Вход' : 'Регистрация'}
        subtitle="Спортивная CRM"
      />
      <div className="flex flex-col gap-3 px-4 pb-8">
        {mode === 'register' && (
          <div className="flex gap-3">
            <input
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              placeholder="Имя"
              className={inputCls}
            />
            <input
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              placeholder="Фамилия"
              className={inputCls}
            />
          </div>
        )}
        <input
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
          placeholder="Телефон, например +79991234567"
          inputMode="tel"
          autoComplete="tel"
          className={inputCls}
        />
        <input
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Пароль"
          type="password"
          autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
          className={inputCls}
          onKeyDown={(e) => e.key === 'Enter' && submit()}
        />

        {error && (
          <div className="rounded-xl bg-red-500/15 p-3 text-sm text-red-300">{error}</div>
        )}

        <Button onClick={submit} disabled={busy}>
          {busy ? 'Подождите...' : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
        </Button>

        <button
          onClick={() => {
            setError('')
            setMode(mode === 'login' ? 'register' : 'login')
          }}
          className="py-1 text-center text-sm text-tg-link"
        >
          {mode === 'login' ? 'Нет аккаунта? Зарегистрироваться' : 'Уже есть аккаунт? Войти'}
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
