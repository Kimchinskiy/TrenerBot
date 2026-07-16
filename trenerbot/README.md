# trenerbot — Mobile First CRM для спортивных тренеров

## Архитектура

Проект состоит из трёх частей:

```
web/                  # Next.js 15 PWA (основной продукт)
trenerbot/            # Go API + Telegram Bot (сервис уведомлений)
nginx/                # reverse proxy (TLS, статика, API)
```

### Web CRM (`web/`)
- Next.js 15 (App Router), React 19, TypeScript
- Tailwind CSS + shadcn/ui
- TanStack Query, React Hook Form + Zod
- PWA (next-pwa)
- Основной способ входа: телефон + пароль (bcrypt)
- Дополнительно: Telegram Login Widget, MAX OAuth (заглушка)
- JWT (Access + Refresh Token)

### Go Backend (`trenerbot/`)
- Chi router, SQLite (modernc.org/sqlite)
- JWT-авторизация, refresh-токены
- Бизнес-логика: clients, coaches, lessons, attendance, notifications, files, reports
- Telegram Bot как сервис уведомлений (polling/webhook)
- Scheduler: напоминания 08:00, outbox-уведомления

### Telegram Bot
- Дополнительный канал уведомлений
- Регистрация через `/start`
- Админ-панель в боте
- Меню с кнопкой открытия Web App

## Запуск

```bash
# Backend
cd trenerbot
cp .env.example .env        # укажите BOT_TOKEN, JWT_SECRET, SERVICE_TOKEN
go mod tidy
go run ./cmd/server &       # API на :8080
go run ./cmd/seed           # создать admin/coach
go run ./cmd/bot            # Telegram-бот (polling)

# Frontend
cd web
cp .env.example .env.local  # укажите NEXT_PUBLIC_TELEGRAM_BOT при необходимости
npm install
npm run dev                 # Next.js на :3000
```

## Переменные окружения

### Backend (`trenerbot/.env`)
| Переменная | Назначение |
|-----------|-----------|
| `HTTP_ADDR` | Адрес API сервера (`:8080`) |
| `DB_PATH` | Путь к SQLite базе (`data/crm.db`) |
| `JWT_SECRET` | Секрет для JWT |
| `SERVICE_TOKEN` | Токен для service-to-service вызовов |
| `BOT_TOKEN` | Telegram Bot token |
| `BOT_MODE` | `polling` или `webhook` |
| `WEBAPP_URL` | URL Web App (для кнопки в боте) |
| `SCHEDULER_INTERVAL` | Интервал проверки уведомлений |
| `ADMIN_TELEGRAM_ID` | Telegram ID администратора |
| `COACH_TELEGRAM_ID` | Telegram ID тренера |

### Frontend (`web/.env.local`)
| Переменная | Назначение |
|-----------|-----------|
| `NEXT_PUBLIC_API_URL` | Базовый URL API (`/api` или полный URL) |
| `NEXT_PUBLIC_TELEGRAM_BOT` | Имя бота для Login Widget (без `@`) |
| `API_BASE_URL` | Target для прокси в dev (`http://localhost:8080`) |

## Стек

- **Frontend:** Next.js 15, React 19, TypeScript, Tailwind CSS, shadcn/ui, TanStack Query, React Hook Form, Zod, next-pwa
- **Backend:** Go 1.26, chi, modernc.org/sqlite, golang-jwt/jwt/v5, go-telegram-bot-api/v5
- **Infrastructure:** nginx (TLS, reverse proxy)

## Статус модулей

- Реализовано: clients, coaches, lessons, attendance, notifications, files, reports, аутентификация (phone/password, Telegram Widget, Mini App initData, MAX OAuth placeholder)
- Фундамент (БД, без логики): groups, subscriptions, payments, freeze, waiting_list, analytics
- В разработке: полноценная логика абонементов/оплат/аналитики

## Деплой

См. `nginx/trenerbot.conf` — пример конфигурации для production.

1. Собрать фронтенд: `npm run build` в `web/`
2. Собрать бэкенд: `go build ./cmd/server` и `go build ./cmd/bot`
3. Настроить nginx для проксирования `/api/*` на `:8080` и статики из `web/.next/`
4. Включить HTTPS (Let's Encrypt)
