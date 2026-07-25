# Startup: Telegram Bot

## Prerequisites

- Go 1.21+
- Backend server running on `localhost:8080` (see STARTUP.md)

## Environment

All config lives in `trenerbot/.env`:

| Variable | Description |
|----------|-------------|
| `BOT_TOKEN` | Telegram Bot API token (from @BotFather) |
| `BOT_MODE` | `polling` (dev) or `webhook` (prod) |
| `API_BASE_URL` | Backend API URL (e.g. `http://localhost:8080`) |
| `SERVICE_TOKEN` | Must match server's `SERVICE_TOKEN` |
| `SCHEDULER_INTERVAL` | Notification poll interval (e.g. `30s`) |
| `WEBAPP_URL` | Optional Mini App URL for the menu button |

## Run

### With logs (interactive)

```powershell
cd trenerbot
go run ./cmd/bot/main.go
```

### In background

```powershell
cd trenerbot
go build -o bot.exe .\cmd\bot\main.go
Start-Process -FilePath ".\bot.exe" -WindowStyle Hidden -WorkingDirectory "$(Get-Location)"
```

### Stop

```powershell
Get-Process -Name "bot" -ErrorAction SilentlyContinue | Stop-Process -Force
```

## Bot commands

| Command | Description |
|---------|-------------|
| `/start` | Register or open main menu |
| `/menu` | Show main menu |
| `/schedule` | Show schedule (select student) |
| `/role <coach|client|admin|clear>` | Debug: override role |

## Registration flow

1. `/start` → enter full name
2. Choose "Себя" or "Ребенка"
3. Enter age, level, phone
4. Lead sent → coach approves via bot
5. Student created → main menu opens

## Architecture

```
Telegram ←→ bot_v2.go (polling/webhook)
                ↓ HTTP (X-Service-Token + X-Telegram-Id)
          Go server (backend)
                ↓ SQLite
          data/crm.db
```

The bot is a stateless HTTP client. All business logic (users, students, leads, trainings, notifications) lives in the Go server. The bot only handles Telegram UI and dispatches API calls.

## Rebuild after changes

```powershell
cd trenerbot
go build -o bot.exe ./cmd/bot/main.go
Get-Process -Name "bot" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Process -FilePath ".\bot.exe" -WindowStyle Hidden -WorkingDirectory "$(Get-Location)"
```
