// services/telegram.ts
// Facade for Telegram-specific integrations.
// Core CRM code should import from here, not from integrations/telegram directly.

export { applyTheme, haptics } from '@/integrations/telegram/theme'
export { showMainButton, hideMainButton } from '@/integrations/telegram/main-button'
export { getInitData, isInsideTelegram, webApp } from '@/integrations/telegram/webapp'
export { TelegramLoginButton } from '@/integrations/telegram/login-widget'
