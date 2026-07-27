# Плавли — Mobile-First CRM для спортивных тренеров

Full-stack CRM для управления спортсменами, расписанием, посещаемостью, абонементами и уведомлениями. Telegram-бот — дополнительный канал для уведомлений и лидогенерации.

Четыре роли: **админ**, **тренер**, **клиент** (спортсмен), **родитель**.

---

## Архитектура

```
trenerbot-tmp/
├── web/            # Next.js 15 PWA (фронтенд)
├── trenerbot/      # Go API сервер + Telegram Bot
├── nginx/          # TLS + reverse proxy (production)
├── data/           # SQLite база (runtime)
└── bin/            # Скомпилированные бинарники
```

### Фронтенд — `web/`

Next.js 15 (App Router), React 19, TypeScript 5.7, Tailwind CSS + shadcn/ui, TanStack Query, React Hook Form + Zod, Framer Motion, PWA (next-pwa).

**Страницы:**
- `/dashboard/home` — главная с быстрыми действиями, расписанием на неделю
- `/dashboard/clients` — управление спортсменами
- `/dashboard/groups` — тренировочные группы
- `/dashboard/schedule` — расписание занятий
- `/dashboard/attendance` — отметка посещаемости по дате
- `/dashboard/leads` — заявки из Telegram
- `/dashboard/statistics` — аналитика (тренировки, клиенты, доход, посещаемость, должники)
- `/dashboard/notifications` — уведомления и рассылки
- `/dashboard/profile` — профиль и настройки
- `/parent/*` — родительский кабинет (привязка к детям, статус занятий)

### Бэкенд — `trenerbot/`

Go 1.26, chi router, SQLite (modernc.org/sqlite, pure Go), JWT (golang-jwt/v5), bcrypt, go-telegram-bot-api/v5.

**Чистая архитектура:** `domain/` → `store/` (SQL) → `service/` → `http/` (handlers).

**Базовые эндпоинты:** аутентификация (phone/password, Telegram Widget, Mini App, MAX OAuth), клиенты, тренеры, занятия, посещаемость, группы, абонементы, статистика, уведомления, заявки, wellbeing-фидбек, файлы.

**Фоновые задачи:**
- Reaper — снимает stale-блокировки уведомлений (каждую минуту)
- Scheduler — ставит напоминания о занятиях в 08:00 (по расписанию)

### Telegram Bot

Polling-режим (или webhook). Регистрация через `/start`, админ-панель, Web App кнопка, доставка уведомлений (напоминания, отмены, рассылки), сбор заявок (leads).

### Инфраструктура — `nginx/`

nginx с Let's Encrypt: прокси Next.js (`:3000`) и Go API (`:8080`), TLS, HTTP→HTTPS редирект.

---

## Аутентификация

Единая модель `users`, несколько способов входа:
1. **Телефон + пароль** (bcrypt) — основной
2. **Telegram Login Widget** — OAuth
3. **Telegram Mini App** — проверка initData
4. **MAX OAuth** — заглушка
5. **Service Token** — для bot→server вызовов

JWT (Access + Refresh Token).

---

## Запуск

```bash
# Backend
cd trenerbot
cp .env.example .env
go mod tidy
go run ./cmd/server &
go run ./cmd/seed               # создать admin + coach
go run ./cmd/bot                # Telegram бот

# Frontend
cd web
cp .env.example .env.local
npm install
npm run dev                     # Next.js на :3000
```

## Переменные окружения

### Backend (`trenerbot/.env`)

| Переменная | Назначение |
|---|---|
| `HTTP_ADDR` | Адрес API (`:8080`) |
| `DB_PATH` | Путь к SQLite (`data/crm.db`) |
| `JWT_SECRET` | Секрет JWT |
| `SERVICE_TOKEN` | Токен для service-to-service |
| `BOT_TOKEN` | Токен Telegram бота |
| `BOT_MODE` | `polling` или `webhook` |
| `WEBAPP_URL` | URL Web App |
| `SCHEDULER_INTERVAL` | Интервал scheduler (напр. `30s`) |
| `ADMIN_TELEGRAM_ID` | Telegram ID админа |
| `COACH_TELEGRAM_ID` | Telegram ID тренера |

### Frontend (`web/.env.local`)

| Переменная | Назначение |
|---|---|
| `NEXT_PUBLIC_API_URL` | Базовый URL API |
| `NEXT_PUBLIC_TELEGRAM_BOT` | Имя бота (без `@`) |
| `API_BASE_URL` | Прокси-таргет в dev (`http://localhost:8080`) |

---

## Статус модулей

✅ Реализовано: аутентификация (phone/password, Telegram, MAX), клиенты, тренеры, занятия, посещаемость, группы, абонементы клиентов, подписка тренера на CRM, заявки (leads), уведомления (outbox + Telegram), wellbeing, аналитика/статистика, родительский кабинет, waiting list, социальные ссылки, FAQ, файлы

> Проект находится в активной разработке.
