package main

import (
	"database/sql"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", os.Args[1])
	if err != nil { panic(err) }
	defer db.Close()

	hash, _ := bcrypt.GenerateFromPassword([]byte("coach123"), bcrypt.DefaultCost)
	
	// Create or update coach user
	var coachUserId int64
	err = db.QueryRow(`SELECT id FROM users WHERE phone = '+79998887766'`).Scan(&coachUserId)
	if err != nil {
		res, err := db.Exec(`INSERT INTO users (phone, password_hash, role) VALUES ('+79998887766', ?, 'coach')`, string(hash))
		if err == nil {
			coachUserId, _ = res.LastInsertId()
		}
	} else {
		_, _ = db.Exec(`UPDATE users SET password_hash = ?, role = 'coach' WHERE id = ?`, string(hash), coachUserId)
	}

	if coachUserId > 0 {
		var coachId int64
		_ = db.QueryRow(`SELECT id FROM coaches WHERE user_id = ?`, coachUserId).Scan(&coachId)
		if coachId == 0 {
			_, _ = db.Exec(`INSERT INTO coaches (user_id, full_name, sport) VALUES (?, 'Сергей Тренеров', 'Плавание')`, coachUserId)
		}
		var clientId int64
		_ = db.QueryRow(`SELECT id FROM clients WHERE user_id = ?`, coachUserId).Scan(&clientId)
		if clientId == 0 {
			_, _ = db.Exec(`INSERT INTO clients (user_id, full_name, phone, status) VALUES (?, 'Сергей Тренеров', '+79998887766', 'active')`, coachUserId)
		}
	}

	// Create or update coach user Алексей
	hashAleksey, _ := bcrypt.GenerateFromPassword([]byte("aleksey"), bcrypt.DefaultCost)
	var alekseyUserId int64
	err = db.QueryRow(`SELECT id FROM users WHERE phone = '+79997923020'`).Scan(&alekseyUserId)
	if err != nil {
		res, err := db.Exec(`INSERT INTO users (phone, password_hash, role) VALUES ('+79997923020', ?, 'coach')`, string(hashAleksey))
		if err == nil {
			alekseyUserId, _ = res.LastInsertId()
		}
	} else {
		_, _ = db.Exec(`UPDATE users SET password_hash = ?, role = 'coach' WHERE id = ?`, string(hashAleksey), alekseyUserId)
	}

	if alekseyUserId > 0 {
		var coachId int64
		_ = db.QueryRow(`SELECT id FROM coaches WHERE user_id = ?`, alekseyUserId).Scan(&coachId)
		if coachId == 0 {
			_, _ = db.Exec(`INSERT INTO coaches (user_id, full_name, sport) VALUES (?, 'Алексей', 'Плавание')`, alekseyUserId)
		}
		var clientId int64
		_ = db.QueryRow(`SELECT id FROM clients WHERE user_id = ?`, alekseyUserId).Scan(&clientId)
		if clientId == 0 {
			_, _ = db.Exec(`INSERT INTO clients (user_id, full_name, phone, status) VALUES (?, 'Алексей', '+79997923020', 'active')`, alekseyUserId)
		}
	}

	fmt.Printf("Coach user 1 (+79998887766) user_id=%d\n", coachUserId)
	fmt.Printf("Coach user 2 (+79997923020) password=aleksey user_id=%d\n", alekseyUserId)
}

