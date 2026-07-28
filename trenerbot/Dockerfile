# --- Build Stage ---
FROM golang:alpine AS builder

WORKDIR /app
ENV GOTOOLCHAIN=auto

# Copy dependency manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bot ./cmd/bot/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o createadmin ./cmd/createadmin/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o seed ./cmd/seed/main.go

# --- Final Stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=UTC

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/server /app/server
COPY --from=builder /app/bot /app/bot
COPY --from=builder /app/createadmin /app/createadmin
COPY --from=builder /app/seed /app/seed

# Create directory for SQLite DB
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["/app/server"]
