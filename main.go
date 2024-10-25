package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	EthereumRPCURL string `yaml:"ethereum_rpc_url"`
	PostgresURL    string `yaml:"postgres_url"`
	Contracts      struct {
		USDC string `yaml:"usdc"`
		ETH  string `yaml:"eth"`
	} `yaml:"contracts"`
}

// Global variable for db connection
var db *sql.DB

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Ethereum node
	ethClient, err := ethclient.Dial(config.EthereumRPCURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum: %v", err)
	}
	defer ethClient.Close()
	fmt.Println("Connected to Ethereum")

	// Connect to PostgreSQL
	db, err = sql.Open("postgres", config.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db, "./migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Connected to PostgreSQL and migrations are done!")

	// Testing Data
	mockEvent := types.Log{
		Address: common.HexToAddress("0xC77A67e42053aa32b483F47794AFCA5c3CAb3595"),
		Topics:  []common.Hash{common.HexToHash("0x8c5be1e5ebec7d5bd14f714f3880a372c2cfde9d4e3c12ed9a5d23b5e3c72da5")}, // Mint event signature
		Data:    common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000032"),                  // Mint amount = 50
	}
	handleEvent(mockEvent, db)

	go listenForEvents(ethClient, db, config)

	http.HandleFunc("/points", getPointsHandler)
	http.ListenAndServe("0.0.0.0:8082", nil)
}

func loadConfig() (*Config, error) {
	config := &Config{}
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func listenForEvents(client *ethclient.Client, db *sql.DB, config *Config) {
	usdcAddress := common.HexToAddress(config.Contracts.USDC)
	ethAddress := common.HexToAddress(config.Contracts.ETH)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{usdcAddress, ethAddress},
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			log.Printf("Error: %v", err)
		case vLog := <-logs:
			handleEvent(vLog, db)
		}
	}
}

func handleEvent(eventLog types.Log, db *sql.DB) {

	log.Printf("Event received with signature: %s", eventLog.Topics[0].Hex())

	eventName := ""
	pointsPerUnit := 0

	if eventLog.Topics[0].Hex() == "0x8c5be1e5ebec7d5bd14f714f3880a372c2cfde9d4e3c12ed9a5d23b5e3c72da5" {
		eventName = "Mint"
		pointsPerUnit = 1
	} else if eventLog.Topics[0].Hex() == "0xc6a898309e1b9f9c7425e34b362b5ecf9b34ad1789b9c235a51d3f4f87b98b21" {
		eventName = "Borrow"
		pointsPerUnit = 2
	} else {
		return
	}

	amount := new(big.Int)
	amount.SetBytes(eventLog.Data[:32])

	duration := 10
	points := pointsPerUnit * int(amount.Int64()) * (duration / 10)

	log.Printf("Amount: %s, Points: %d", amount.String(), points)

	_, err := db.Exec("INSERT INTO user_points (address, points, event) VALUES ($1, $2, $3)",
		eventLog.Address.Hex(), points, eventName)
	if err != nil {
		fmt.Printf("Failed to insert points: %v", err)
	}
}

func getPointsHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	var points int
	err := db.QueryRow("SELECT SUM(points) FROM user_points WHERE address = $1", address).Scan(&points)
	if err != nil {
		http.Error(w, "Failed to retrieve points", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Address: %s, Points: %d", address, points)
}

func runMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)

	if err != nil {
		return err
	}

	// if err := m.Down(); err != nil && err != migrate.ErrNoChange {
	// 	return err
	// }

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations applied successfully!")
	return nil
}
