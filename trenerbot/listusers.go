package main

import (
	"database/sql"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func upsertUser(db *sql.DB, phone, password, role, name string) int64 {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Bcrypt error for %s: %v\n", phone, err)
		return 0
	}

	var userID int64
	err = db.QueryRow(`SELECT id FROM users WHERE phone = ?`, phone).Scan(&userID)
	if err != nil {
		// Not found by phone, check if there's an unassigned user with matching role without phone
		err2 := db.QueryRow(`SELECT id FROM users WHERE role = ? AND (phone IS NULL OR phone = '')`, role).Scan(&userID)
		if err2 == nil && userID > 0 {
			_, err = db.Exec(`UPDATE users SET phone = ?, password_hash = ?, first_name = ?, updated_at = datetime('now') WHERE id = ?`, phone, string(hash), name, userID)
			if err != nil {
				fmt.Printf("Update error for %s: %v\n", phone, err)
			}
		} else {
			res, err := db.Exec(`INSERT INTO users (phone, password_hash, role, first_name, updated_at) VALUES (?, ?, ?, ?, datetime('now'))`, phone, string(hash), role, name)
			if err != nil {
				fmt.Printf("Insert error for %s: %v\n", phone, err)
				return 0
			}
			userID, _ = res.LastInsertId()
		}
	} else {
		_, err = db.Exec(`UPDATE users SET password_hash = ?, role = ?, first_name = ?, updated_at = datetime('now') WHERE id = ?`, string(hash), role, name, userID)
		if err != nil {
			fmt.Printf("Update error for %s: %v\n", phone, err)
		}
	}

	if userID > 0 {
		if role == "coach" || role == "admin" {
			var coachID int64
			_ = db.QueryRow(`SELECT id FROM coaches WHERE user_id = ?`, userID).Scan(&coachID)
			if coachID == 0 {
				_, _ = db.Exec(`INSERT INTO coaches (user_id, full_name, sport) VALUES (?, ?, 'Плавание')`, userID, name)
			}
		}
		var clientID int64
		_ = db.QueryRow(`SELECT id FROM clients WHERE user_id = ?`, userID).Scan(&clientID)
		if clientID == 0 {
			_, _ = db.Exec(`INSERT INTO clients (user_id, full_name, phone, status) VALUES (?, ?, ?, 'active')`, userID, name, phone)
		}
	}

	return userID
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run listusers.go <path-to-db>")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", os.Args[1])
	if err != nil {
		panic(err)
	}
	defer db.Close()

	idAdmin := upsertUser(db, "+79999999999", "admin123", "admin", "Администратор")
	idAleksey := upsertUser(db, "+79997923020", "aleksey", "coach", "Алексей")
	idSergey := upsertUser(db, "+79998887766", "coach123", "coach", "Сергей Тренеров")

	fmt.Printf("✅ Admin (+79999999999 / admin123) user_id=%d\n", idAdmin)
	fmt.Printf("✅ Coach Aleksey (+79997923020 / aleksey) user_id=%d\n", idAleksey)
	fmt.Printf("✅ Coach Sergey (+79998887766 / coach123) user_id=%d\n", idSergey)
}
