// integrations/telegram/webapp.ts
// Accessors for Telegram WebApp SDK and initData.

function webApp(): any {
  return (window as any).Telegram?.WebApp
}

export function getInitData(): string {
  return webApp()?.initData ?? ''
}

export function isInsideTelegram(): boolean {
  return !!getInitData()
}

export { webApp }
