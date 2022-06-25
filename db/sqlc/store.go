package db

import (
	"database/sql"
)

type Store interface {
	Querier
}

//Store provides all functions to execute db queries
type SQLStore struct {
	*Queries
	db *sql.DB
}

//NewStore create a new Store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}
