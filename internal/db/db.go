package db

import 
(
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open() (*DB, error) {
	homepath, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}

	dbdir := filepath.Join(homepath, ".dotfiles", "dashboard")
	if err := os.MkdirAll(dbdir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	dbpath := filepath.Join(dbdir, "database.db")

	db, err := sql.Open("sqlite", dbpath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &DB{db}, nil
}
