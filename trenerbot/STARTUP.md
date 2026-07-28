# Startup: Server + Frontend

## Prerequisites

- Go 1.21+
- Node.js 20+
- npm

## 1. Backend Server (Go)

### Prepare environment

Edit `trenerbot/.env`:

```env
HTTP_ADDR=:8080
DB_PATH=data/crm.db
JWT_SECRET=dev-secret-change-in-prod
SERVICE_TOKEN=dev-token
BOT_TOKEN=<your-telegram-bot-token>
BOT_MODE=polling
API_BASE_URL=http://localhost:8080
SCHEDULER_INTERVAL=30s
```

### Run

```powershell
cd trenerbot
go build -o server.exe ./cmd/server/main.go
Start-Process -FilePath ".\server.exe" -WindowStyle Hidden -WorkingDirectory "$(Get-Location)"
```

Verify: `curl http://localhost:8080/health` → `{"status":"ok"}`

## 2. Frontend (Next.js)

### Prepare environment

Create `web/.env`:

```env
NEXT_PUBLIC_API_URL=/api
NEXT_PUBLIC_TELEGRAM_BOT=
API_BASE_URL=http://localhost:8080
```

### Run

```powershell
cd web
npm install
Start-Process -FilePath "cmd" -ArgumentList "/c npx next dev --port 3000" -WindowStyle Hidden -WorkingDirectory "$(Get-Location)"
```

Verify: open http://localhost:3000 in browser.

## Quick start (both)

```powershell
# Terminal 1 — server
cd trenerbot
go run ./cmd/server/main.go

# Terminal 2 — frontend
cd web
npm run dev
```

## Useful commands

| Command | Description |
|---------|-------------|
| `go run ./cmd/server/main.go` | Run server with logs |
| `go run ./cmd/seed/main.go` | Seed test data |
| `npm run dev` | Start Next.js dev server |
| `go build ./cmd/server/main.go` | Build server binary |
