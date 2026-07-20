package main

import (
	"fmt"
	"log/slog"
	"os"

	"trenerbot/internal/auth"
	"trenerbot/internal/db"
	"trenerbot/internal/store"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/crm.db"
	}
	database, err := db.Open(dbPath)
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer database.Close()
	s := store.New(database)

	phone := "+79999999999"
	password := "admin123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		slog.Error("hash", "err", err)
		os.Exit(1)
	}

	existing, _ := s.UserByPhone(phone)
	if existing != nil {
		fmt.Println("Admin already exists with phone:", phone)
		fmt.Printf("Phone: %s\nPassword: %s\n", phone, password)
		return
	}

	// Update the existing admin (by telegram) with phone+password
	_, err = s.DB.Exec(`UPDATE users SET phone=?, password_hash=?, updated_at=datetime('now') WHERE telegram_id=? AND role='admin'`, phone, hash, "378450978")
	if err != nil {
		slog.Error("update admin", "err", err)
		os.Exit(1)
	}

	fmt.Printf("Admin ready\nPhone: %s\nPassword: %s\n", phone, password)
}
