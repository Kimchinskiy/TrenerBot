package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens (creating if needed) the SQLite database and applies pending migrations.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir db dir: %w", err)
		}
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := d.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := Migrate(d); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

func Migrate(d *sql.DB) error {
	if _, err := d.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	type mig struct{ version, body string }
	var migs []mig
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		b, err := fs.ReadFile(migrationsFS, "migrations/"+e.Name())
		if err != nil {
			return err
		}
		migs = append(migs, mig{version: e.Name(), body: string(b)})
	}
	sort.Slice(migs, func(i, j int) bool { return migs[i].version < migs[j].version })

	for _, m := range migs {
		var exists int
		if err := d.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, m.version).Scan(&exists); err != nil && err != sql.ErrNoRows {
			return err
		}
		if exists == 1 {
			continue
		}
		if _, err := d.Exec(m.body); err != nil {
			return fmt.Errorf("apply migration %s: %w", m.version, err)
		}
		if _, err := d.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, m.version); err != nil {
			return err
		}
	}
	return nil
}
