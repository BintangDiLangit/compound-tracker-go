package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/BintangDiLangit/compound-tracker/internal/api"
	"github.com/BintangDiLangit/compound-tracker/internal/config"
	"github.com/BintangDiLangit/compound-tracker/internal/db"
	"github.com/BintangDiLangit/compound-tracker/internal/events"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
)

func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}

func main() {
	// Load config
	cfg, err := config.Load(filepath.Join(getProjectRoot(), "config.yaml"))
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
	database, err := db.Connect(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.RunMigrations(database, "./migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Start event listener
	go events.ListenForEvents(ethClient, database, cfg)

	// Start HTTP server
	http.HandleFunc("/points", api.GetPointsHandler(database))
	http.ListenAndServe("0.0.0.0:8082", nil)
}
