package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func Connect(postgresURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return nil, err
	}
	return db, nil
}
