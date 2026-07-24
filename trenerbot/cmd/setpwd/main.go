package main

import (
	"fmt"
	"log"

	"trenerbot/internal/auth"
	"trenerbot/internal/db"
)

func main() {
	database, err := db.Open("data/crm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	pwd := "123456"
	var exists int
	if err := database.QueryRow("SELECT 1 FROM users WHERE id = 1").Scan(&exists); err != nil {
		log.Fatalf("User 1 not found: %v", err)
	}

	hash, err := auth.HashPassword(pwd)
	if err != nil {
		log.Fatal(err)
	}

	_, err = database.Exec(`UPDATE users SET phone = '+70000000001', password_hash = ? WHERE id = 1`, hash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("User 1 updated: phone=+70000000001, password=123456")
	fmt.Println()
	fmt.Println("Вход в приложение:")
	fmt.Println("  Телефон: +70000000001")
	fmt.Println("  Пароль: 123456")
}
