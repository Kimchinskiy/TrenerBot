package main

import (
	"log/slog"
	"os"

	"trenerbot/internal/bot"
	"trenerbot/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	b, err := bot.New(cfg)
	if err != nil {
		slog.Error("bot init", "err", err)
		os.Exit(1)
	}
	if err := b.Start(); err != nil {
		slog.Error("bot", "err", err)
		os.Exit(1)
	}
}
