// Applies Telegram theme parameters as CSS variables and expands the viewport,
// so the Mini App looks native. Falls back to defaults when not inside Telegram.
export function applyTheme() {
  const tw: any = (window as any).Telegram?.WebApp
  if (!tw) return

  tw.ready?.()
  tw.expand?.()

  const p = tw.themeParams || {}
  const root = document.documentElement
  const set = (cssVar: string, value?: string) => {
    if (value) root.style.setProperty(cssVar, value)
  }
  set('--tg-theme-bg-color', p.bg_color)
  set('--tg-theme-secondary-bg-color', p.secondary_bg_color)
  set('--tg-theme-text-color', p.text_color)
  set('--tg-theme-hint-color', p.hint_color)
  set('--tg-theme-link-color', p.link_color)
  set('--tg-theme-button-color', p.button_color)
  set('--tg-theme-button-text-color', p.button_text_color)

  // Match the status bar color to the background.
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta && p.bg_color) meta.setAttribute('content', p.bg_color)
}

export const haptics = {
  light() {
    ;(window as any).Telegram?.WebApp?.HapticFeedback?.impactOccurred?.('light')
  },
  medium() {
    ;(window as any).Telegram?.WebApp?.HapticFeedback?.impactOccurred?.('medium')
  },
  success() {
    ;(window as any).Telegram?.WebApp?.HapticFeedback?.notificationOccurred?.('success')
  },
  error() {
    ;(window as any).Telegram?.WebApp?.HapticFeedback?.notificationOccurred?.('error')
  },
}
