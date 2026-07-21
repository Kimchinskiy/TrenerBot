package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "data/crm.db")
	if err != nil {
		fmt.Println("db error:", err)
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, phone, role, first_name, last_name FROM users WHERE role = 'admin' OR role = 'coach'`)
	if err != nil {
		fmt.Println("query error:", err)
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var phone, role, first, last string
		if err := rows.Scan(&id, &phone, &role, &first, &last); err != nil {
			fmt.Println("scan error:", err)
			continue
		}
		fmt.Printf("%d | %s | %s | %s %s\n", id, role, phone, first, last)
	}
}
