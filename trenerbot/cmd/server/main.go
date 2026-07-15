package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"trenerbot/internal/auth"
	"trenerbot/internal/config"
	"trenerbot/internal/db"
	httppkg "trenerbot/internal/http"
	"trenerbot/internal/service"
	"trenerbot/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	_ = os.MkdirAll("data", 0o755)

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	s := store.New(database)
	tokens := auth.NewTokenService(cfg.JWTSecret, 24*time.Hour)
	svc := service.New(s, tokens)

	// Background jobs: reap stale claims + ensure 08:00 reminders (ТЗ §9/§15).
	go func() {
		t := time.NewTicker(time.Minute)
		for range t.C {
			if n, err := svc.ReapStaleClaimed(10 * time.Minute); err == nil && n > 0 {
				slog.Info("reaped stale notifications", "n", n)
			}
		}
	}()
	go func() {
		t := time.NewTicker(cfg.SchedulerInterval)
		for range t.C {
			if n, err := svc.EnsureReminders(); err != nil {
				slog.Error("ensure reminders", "err", err)
			} else if n > 0 {
				slog.Info("enqueued reminders", "n", n)
			}
		}
	}()

	r := httppkg.Router(svc, cfg)
	slog.Info("server listening", "addr", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
