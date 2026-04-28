package sqlite

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(storagePath string) (*Storage, error) {
	const operation = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	tx, err := db.Begin() // транзакция нужна для создания нескольких таблиц без отдельных Prepare
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer tx.Rollback()

	queries := []string{`
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_alias ON urls(alias);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}
