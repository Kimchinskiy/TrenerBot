export function applyTheme() {
  const tw = (window as any).Telegram?.WebApp
  if (tw) {
    tw.ready?.()
    tw.expand?.()
  }
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
