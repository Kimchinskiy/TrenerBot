// integrations/telegram/main-button.ts
// Thin wrapper around the Telegram MainButton.

type MainButtonHandler = () => void

export function showMainButton(text: string, onClick: MainButtonHandler) {
  const mb = (window as any).Telegram?.WebApp?.MainButton
  if (!mb) return
  mb.setParams?.({
    text,
    color:
      getComputedStyle(document.documentElement).getPropertyValue('--tg-theme-button-color').trim() ||
      '#2481cc',
  })
  mb.show?.()
  mb.offClick?.(showMainButton._handler)
  showMainButton._handler = onClick
  mb.onClick?.(onClick)
}

showMainButton._handler = undefined as MainButtonHandler | undefined

export function hideMainButton() {
  const mb = (window as any).Telegram?.WebApp?.MainButton
  if (!mb) return
  if (showMainButton._handler) mb.offClick?.(showMainButton._handler)
  showMainButton._handler = undefined
  mb.hide?.()
}
