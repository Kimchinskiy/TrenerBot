# trenerbot — CRM для спортивных тренеров

CRM-ядро (Go + SQLite) с Telegram-ботом как основным интерфейсом. Бот — тонкий адаптер,
вся логика живёт в backend API (ТЗ §2/§15).

## Архитектура
```
Telegram ──> cmd/bot (адаптер) ──HTTP──> cmd/server (chi + service + sqlite)
                                          └── scheduler: напоминания 08:00, уведомления
```
- Бот ходит в backend только по REST API (заголовки `X-Service-Token` + `X-Telegram-Id`).
- Уведомления: outbox-таблица `notifications` с БД-блокировкой (`status` + `claim_token`) → без дублей при нескольких инстансах.

## Запуск
```bash
cp .env.example .env        # укажите BOT_TOKEN и SERVICE_TOKEN
go mod tidy
go run ./cmd/server &       # backend на :8080
go run ./cmd/seed           # создать admin/coach (задайте ADMIN/COACH_TELEGRAM_ID в .env)
go run ./cmd/bot            # Telegram-бот (polling)
```

## Сценарии бота
- **Клиент:** `/start` → регистрация (ФИО, телефон, возраст, мед.ограничения) → «Моё расписание», «Связь с тренером».
- **Родитель:** видит расписание детей.
- **Тренер:** «Моё расписание», «Отметить посещаемость» (✓/✗), «Загрузить фото».

Новый клиент → тренер получает уведомление «Новый клиент». Расписание → напоминание в 08:00.

## Статус модулей (итерация 1)
- Реализовано: clients, coaches, lessons, attendance, notifications, files, reports, аутентификация.
- Фундамент (БД, без логики): groups, subscriptions, payments, freeze, waiting_list, analytics.
- Отложено: React-панель, MAX/WhatsApp, Redis, полная логика абонементов/оплат/аналитики, AI.

## Стек
Go 1.26 · chi · modernc.org/sqlite · golang-jwt · go-telegram-bot-api/v5 · swaggo (Swagger — след. итерация).
