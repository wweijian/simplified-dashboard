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
		fmt.Errorf("%w", err)
		return nil, err
	}
	
	dbpath := filepath.Join(homepath, ".dotfiles/dashboard/database.db")
	db, err := sql.Open("sqlite", dbpath);
	if err != nil {
		return nil, fmt.Errorf("could not open: %w", err)
	}
	
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &DB{db}, nil
}
