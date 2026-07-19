// integrations/telegram/theme.ts
// Theme and haptic feedback adapters for Telegram WebApp.

function parseHex(hex: string) {
  hex = hex.replace('#', '')
  if (hex.length === 3) {
    hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2]
  }
  if (hex.length !== 6) return null
  return {
    r: parseInt(hex.substring(0, 2), 16) / 255,
    g: parseInt(hex.substring(2, 4), 16) / 255,
    b: parseInt(hex.substring(4, 6), 16) / 255,
  }
}

function hexToHsl(hex: string, adjustL: number = 0): string | null {
  const rgb = parseHex(hex)
  if (!rgb) return null

  const { r, g, b } = rgb
  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  let h = 0
  let s = 0
  const l = (max + min) / 2

  if (max !== min) {
    const d = max - min
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
    switch (max) {
      case r:
        h = (g - b) / d + (g < b ? 6 : 0)
        break
      case g:
        h = (b - r) / d + 2
        break
      case b:
        h = (r - g) / d + 4
        break
    }
    h /= 6
  }

  let finalL = Math.round(l * 100) + adjustL
  if (finalL > 100) finalL = 100
  if (finalL < 0) finalL = 0

  return `${Math.round(h * 360)} ${Math.round(s * 100)}% ${finalL}%`
}

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

  // Telegram core theme colors
  set('--tg-theme-bg-color', p.bg_color)
  set('--tg-theme-secondary-bg-color', p.secondary_bg_color)
  set('--tg-theme-text-color', p.text_color)
  set('--tg-theme-hint-color', p.hint_color)
  set('--tg-theme-link-color', p.link_color)
  set('--tg-theme-button-color', p.button_color)
  set('--tg-theme-button-text-color', p.button_text_color)

  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta && p.bg_color) meta.setAttribute('content', p.bg_color)

  // Map theme colors to shadcn HSL CSS variables
  const setHsl = (shadcnVar: string, hex?: string, adjustL: number = 0) => {
    if (!hex) return
    const hsl = hexToHsl(hex, adjustL)
    if (hsl) root.style.setProperty(shadcnVar, hsl)
  }

  const bgHex = p.bg_color || '#0d0d0d'
  const secHex = p.secondary_bg_color || '#1c1c1e'
  const txtHex = p.text_color || '#ffffff'
  const btnHex = p.button_color || '#2481cc'
  const btnTxtHex = p.button_text_color || '#ffffff'
  const hintHex = p.hint_color || '#999999'

  // Determine if theme is light or dark
  const rgb = parseHex(bgHex)
  const isDark = rgb ? (0.299 * rgb.r + 0.587 * rgb.g + 0.114 * rgb.b) < 0.5 : true
  const borderShift = isDark ? 8 : -8
  const mutedShift = isDark ? -2 : 2

  setHsl('--background', bgHex)
  setHsl('--foreground', txtHex)
  setHsl('--card', secHex)
  setHsl('--card-foreground', txtHex)
  setHsl('--popover', secHex)
  setHsl('--popover-foreground', txtHex)
  setHsl('--primary', btnHex)
  setHsl('--primary-foreground', btnTxtHex)
  setHsl('--secondary', secHex)
  setHsl('--secondary-foreground', txtHex)
  setHsl('--muted', secHex, mutedShift)
  setHsl('--muted-foreground', hintHex)
  setHsl('--accent', secHex, borderShift)
  setHsl('--accent-foreground', txtHex)
  setHsl('--border', secHex, borderShift)
  setHsl('--input', secHex, borderShift)
  setHsl('--ring', btnHex)
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

