package events

import (
	"context"
	"database/sql"
	"log"
	"time"

	"math/big"

	"github.com/BintangDiLangit/compound-tracker/internal/config"
	"github.com/BintangDiLangit/compound-tracker/internal/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ListenForEvents(client *ethclient.Client, db *sql.DB, cfg *config.Config) {

	latestBlockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch latest block number: %v", err)
	}

	if err != nil {
		log.Fatalf("Failed to fetch latest block header: %v", err)
	}

	startBlock := int64(latestBlockNumber)

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(startBlock),
		ToBlock:   nil,
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Listening for events...")

	for {
		select {
		case err := <-sub.Err():
			log.Printf("Subscription error: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			sub, _ = client.SubscribeFilterLogs(context.Background(), query, logs)
		case vLog := <-logs:
			HandleEvent(vLog, client, db, cfg)
			utils.SaveLastProcessedBlock(int64(vLog.BlockNumber))
		}
	}
}
