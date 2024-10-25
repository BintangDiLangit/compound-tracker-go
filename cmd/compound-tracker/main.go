package main

import (
	"compound-tracker/internal/api"
	"compound-tracker/internal/config"
	"compound-tracker/internal/db"
	"compound-tracker/internal/events"
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Ethereum node
	ethClient, err := ethclient.Dial(cfg.EthereumRPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum: %v", err)
	}
	defer ethClient.Close()
	fmt.Println("Connected to Ethereum")

	// Connect to PostgreSQL
	dbConn, err := db.Connect(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer dbConn.Close()

	// Run migrations
	if err := db.RunMigrations(dbConn, "./migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Start event listener
	go events.ListenForEvents(ethClient, dbConn, cfg)

	// Start HTTP server
	http.HandleFunc("/points", api.GetPointsHandler(dbConn))
	http.ListenAndServe("0.0.0.0:8082", nil)
}
