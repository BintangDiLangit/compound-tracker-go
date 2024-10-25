package events

import (
	"context"
	"database/sql"
	"log"
	"math/big"

	"compound-tracker/internal/config"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ListenForEvents(client *ethclient.Client, db *sql.DB, cfg *config.Config) {
	usdcAddress := common.HexToAddress(cfg.Contracts.USDC)
	ethAddress := common.HexToAddress(cfg.Contracts.ETH)

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

	switch eventLog.Topics[0].Hex() {
	case "0x8c5be1e5ebec7d5bd14f714f3880a372c2cfde9d4e3c12ed9a5d23b5e3c72da5":
		eventName = "Mint"
		pointsPerUnit = 1
	case "0xc6a898309e1b9f9c7425e34b362b5ecf9b34ad1789b9c235a51d3f4f87b98b21":
		eventName = "Borrow"
		pointsPerUnit = 2
	default:
		return
	}

	amount := new(big.Int).SetBytes(eventLog.Data[:32])
	points := pointsPerUnit * int(amount.Int64()) * (10 / 10)

	log.Printf("Amount: %s, Points: %d", amount.String(), points)

	_, err := db.Exec("INSERT INTO user_points (address, points, event) VALUES ($1, $2, $3)",
		eventLog.Address.Hex(), points, eventName)
	if err != nil {
		log.Printf("Failed to insert points: %v", err)
	}
}
