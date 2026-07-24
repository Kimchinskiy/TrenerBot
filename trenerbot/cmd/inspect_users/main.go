package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "data/crm.db")
	if err != nil {
		slog.Error("db error", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, phone, role, first_name, last_name FROM users WHERE role = 'admin' OR role = 'coach'`)
	if err != nil {
		slog.Error("query error", "err", err)
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var phone, role, first, last string
		if err := rows.Scan(&id, &phone, &role, &first, &last); err != nil {
			slog.Error("scan error", "err", err)
			continue
		}
		slog.Info("user", "id", id, "role", role, "phone", phone, "first_name", first, "last_name", last)
	}
}
