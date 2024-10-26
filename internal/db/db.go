package db

import (
	"database/sql"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	_ "github.com/lib/pq"
)

var database *sql.DB

func Connect(postgresURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database = db

	log.Println("Successfully connected to database")
	return db, nil
}

func InsertUserPoints(db *sql.DB, eventLog types.Log, points int64, eventName string) {
	if db == nil {
		log.Println("Database connection is nil")
		return
	}

	_, err := db.Exec(`
		INSERT INTO user_points 
		(address, points, event, block_number, transaction_hash, log_index, amount) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		eventLog.Address.Hex(),
		points,
		eventName,
		eventLog.BlockNumber,
		eventLog.TxHash.Hex(),
		eventLog.Index,
		"0")
	if err != nil {
		log.Printf("Failed to insert points: %v", err)
	}
}

func Close() {
	if database != nil {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}
}

func GetDB() *sql.DB {
	return database
}
