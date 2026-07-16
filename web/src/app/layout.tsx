import type { Metadata, Viewport } from 'next'
import './globals.css'
import { Providers } from '@/components/providers'
import { AuthProvider } from '@/components/auth-provider'
import { ApplyTheme } from '@/components/apply-theme'
import { InstallPrompt } from '@/components/install-prompt'

export const metadata: Metadata = {
  title: 'Плавли — Спортивная CRM',
  description: 'Мобильное приложение тренера и клиента',
  manifest: '/manifest.webmanifest',
  appleWebApp: {
    capable: true,
    statusBarStyle: 'black-translucent',
    title: 'Плавли',
  },
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
  viewportFit: 'cover',
  themeColor: '#0d0d0d',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body>
        <Providers>
          <AuthProvider>
            <ApplyTheme />
            {children}
            <InstallPrompt />
          </AuthProvider>
        </Providers>
      </body>
    </html>
  )
}
