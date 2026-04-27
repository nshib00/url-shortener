package sqlite

import (
	"errors"
	"fmt"
	"go-url-shortener/internal/hashpwd"
	"go-url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

func (s *Storage) CreateUser(username string, password string) (int64, error) {
	const operation = "storage.sqlite.CreateUser"

	hash, err := hashpwd.HashPassword(password)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	stmt, err := s.db.Prepare("INSERT INTO users (name, password_hash) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, hash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return 0, fmt.Errorf("%s: %w", operation, storage.ErrUserAlreadyExists)
			}
		}
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return id, nil
}
