package finance

import "simplified-dashboard/internal/db"

type Store struct {
	db *db.DB
}

func NewStore(database *db.DB) *Store {
	return &Store{db: database}
}
