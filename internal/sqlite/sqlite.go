package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	DB *sql.DB
}

func New() (*SQLite, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		`
			CREATE TABLE IF NOT EXISTS content (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				slug TEXT NOT NULL,
				title TEXT NOT NULL,
				created_at DATETIME NOT NULL,
				updated_at DATETIME DEFAULT NULL,
				metadata JSONB DEFAULT NULL
			)
		`,
	)
	if err != nil {
		return nil, err
	}

	return &SQLite{
		DB: db,
	}, nil
}
