package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"simplified-dashboard/internal/appenv"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func Open() (*DB, error) {
	dbpath, err := PathFromEnv()
	if err != nil {
		return nil, err
	}
	return OpenPath(dbpath)
}

func PathFromEnv() (string, error) {
	path := os.Getenv("DASHBOARD_DB_PATH")
	if path == "" {
		return "", fmt.Errorf("DASHBOARD_DB_PATH is required")
	}
	return appenv.ExpandPath(path), nil
}

func OpenPath(dbpath string) (*DB, error) {
	dbpath = appenv.ExpandPath(dbpath)
	dbdir := filepath.Dir(dbpath)
	if err := os.MkdirAll(dbdir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbpath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	if err := migrateHabitLogStatus(db); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{db}, nil
}

func migrateHabitLogStatus(db *sql.DB) error {
	hasStatus, err := tableHasColumn(db, "habit_logs", "status")
	if err != nil {
		return fmt.Errorf("inspect habit logs schema: %w", err)
	}
	if hasStatus {
		return nil
	}

	if _, err := db.Exec(`
		ALTER TABLE habit_logs
		ADD COLUMN status TEXT NOT NULL DEFAULT 'incomplete'
	`); err != nil {
		return fmt.Errorf("add habit log status column: %w", err)
	}

	if _, err := db.Exec(`
		UPDATE habit_logs
		SET status = CASE completed WHEN 1 THEN 'complete' ELSE 'incomplete' END
	`); err != nil {
		return fmt.Errorf("backfill habit log status: %w", err)
	}

	return nil
}

func tableHasColumn(db *sql.DB, tableName, columnName string) (bool, error) {
	rows, err := db.Query(`PRAGMA table_info(` + tableName + `)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, err
	}
	return false, nil
}
