package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", os.Args[1])
	if err != nil { panic(err) }
	defer db.Close()
	rows, err := db.Query(`SELECT id, phone, password_hash, telegram_id, role FROM users`)
	if err != nil { panic(err) }
	defer rows.Close()
	fmt.Println("id | phone | password_hash | telegram_id | role")
	for rows.Next() {
		var id int64; var phone, ph, tg, role sql.NullString
		rows.Scan(&id, &phone, &ph, &tg, &role)
		fmt.Printf("%d | %v | %v | %v | %v\n", id, phone, ph, tg, role)
	}
}
