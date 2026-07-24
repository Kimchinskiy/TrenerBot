import type { Metadata, Viewport } from 'next'
import './globals.css'
import { Providers } from '@/components/providers'
import { AuthProvider } from '@/components/auth-provider'
import { ApplyTheme } from '@/components/apply-theme'
export const metadata: Metadata = {
  title: 'Плавли',
  description: 'Веб-приложение тренера и клиента',
  manifest: '/manifest.webmanifest',
  icons: [{ rel: 'icon', url: '/icon3.svg', type: 'image/svg+xml' }],
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
          </AuthProvider>
        </Providers>
      </body>
    </html>
  )
}
