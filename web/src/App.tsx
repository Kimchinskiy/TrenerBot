import { useEffect, useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { login } from './lib/auth'
import { applyTheme } from './lib/theme'
import { useMe } from './lib/hooks'
import { Spinner } from './components/ui'
import type { Role } from './lib/types'

import Schedule from './screens/Schedule'
import Profile from './screens/Profile'
import More, { type ExtraScreen } from './screens/More'
import Attendance from './screens/Attendance'
import Wellbeing from './screens/Wellbeing'
import Contact from './screens/Contact'
import Debtors from './screens/Debtors'
import WaitingList from './screens/WaitingList'
import Social from './screens/Social'

type View = 'schedule' | 'profile' | 'more' | ExtraScreen

const EXTRA: ExtraScreen[] = ['attendance', 'wellbeing', 'contact', 'debtors', 'waiting', 'social']

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } },
})

function Tab({ active, label, onClick }: { active: boolean; label: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`flex-1 py-3 text-sm font-medium ${
        active ? 'text-tg-button' : 'text-tg-hint'
      }`}
    >
      {label}
    </button>
  )
}

function Shell() {
  const [view, setView] = useState<View>('schedule')
  const me = useMe()
  const role: Role = me.data?.role ?? 'client'

  const goExtra = (s: ExtraScreen) => setView(s)
  const goBack = () => setView('more')

  const render = () => {
    switch (view) {
      case 'schedule':
        return <Schedule />
      case 'profile':
        return <Profile />
      case 'more':
        return <More role={role} onNavigate={goExtra} />
      case 'attendance':
        return <Attendance onBack={goBack} />
      case 'wellbeing':
        return <Wellbeing onBack={goBack} />
      case 'contact':
        return <Contact onBack={goBack} />
      case 'debtors':
        return <Debtors onBack={goBack} />
      case 'waiting':
        return <WaitingList onBack={goBack} />
      case 'social':
        return <Social onBack={goBack} />
    }
  }

  const tab: View = EXTRA.includes(view as ExtraScreen) ? 'more' : view

  return (
    <div className="flex h-full flex-col">
      <div className="flex-1 overflow-y-auto">{render()}</div>
      <nav className="flex border-t border-tg-secondary bg-tg-bg">
        <Tab active={tab === 'schedule'} label="📅 Расписание" onClick={() => setView('schedule')} />
        <Tab active={tab === 'profile'} label="👤 Профиль" onClick={() => setView('profile')} />
        <Tab active={tab === 'more'} label="☰ Ещё" onClick={() => setView('more')} />
      </nav>
    </div>
  )
}

function App() {
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading')
  const [authError, setAuthError] = useState('')

  useEffect(() => {
    applyTheme()
    login()
      .then(() => setStatus('ready'))
      .catch((e: unknown) => {
        setAuthError(e instanceof Error ? e.message : 'Ошибка входа')
        setStatus('error')
      })
  }, [])

  if (status === 'loading') {
    return (
      <div className="flex h-full items-center justify-center">
        <Spinner label="Вход..." />
      </div>
    )
  }
  if (status === 'error') {
    return <div className="p-6 text-center text-tg-hint">{authError}</div>
  }
  return <Shell />
}

export default function Root() {
  return (
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  )
}
