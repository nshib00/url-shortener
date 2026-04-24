package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"go-url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const operation = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO urls (alias, url) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	result, err := stmt.Exec(alias, urlToSave)
	if err != nil {
		if sqliteErr, ok := errors.AsType[sqlite3.Error](err); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return 0, fmt.Errorf("%s: %w", operation, storage.ErrURLAlreadyExists)
			}
		}
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const operation = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM urls WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", operation, err)
	}

	var url string
	err = stmt.QueryRow(alias).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", operation, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", operation, err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const operation = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM urls WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	result, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: %w", operation, storage.ErrURLNotFound)
	}

	return nil
}
