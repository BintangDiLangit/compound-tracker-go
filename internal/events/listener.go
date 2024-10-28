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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ListenForEvents(client *ethclient.Client, db *sql.DB, cfg *config.Config) {
	currentLatestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch latest block number: %v", err)
	}

	lastProcessedBlock := utils.GetLastProcessedBlock()
	currentBlock := int64(lastProcessedBlock)
	if currentBlock == 0 {
		currentBlock = 21059388
	}

	// Mulai dengan batch yang sangat kecil
	batchSize := int64(10)
	maxRetries := 3

	log.Printf("Starting scan from block %d to %d", currentBlock, currentLatestBlock)

	for currentBlock < int64(currentLatestBlock) {
		endBlock := currentBlock + batchSize
		if endBlock > int64(currentLatestBlock) {
			endBlock = int64(currentLatestBlock)
		}

		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(currentBlock),
			ToBlock:   big.NewInt(endBlock),
		}

		retryCount := 0
		success := false

		for !success && retryCount < maxRetries {
			log.Printf("Scanning blocks %d to %d", currentBlock, endBlock)

			logs, err := client.FilterLogs(context.Background(), query)
			if err != nil {
				retryCount++
				log.Printf("Error (attempt %d/%d): %v", retryCount, maxRetries, err)
				time.Sleep(time.Second)
				continue
			}

			for _, vLog := range logs {
				HandleEvent(vLog, client, db, cfg)
				utils.SaveLastProcessedBlock(int64(vLog.BlockNumber))
			}
			success = true
		}

		if !success {
			log.Printf("Skipping problematic block range after %d retries", maxRetries)
		}

		currentBlock = endBlock + 1
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Scan completed at block %d", currentLatestBlock)
}

func listenForNewEvents(client *ethclient.Client, db *sql.DB, cfg *config.Config, startBlock int64) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   nil,
		Addresses: []common.Address{
			common.HexToAddress(cfg.Contracts.ETH),
		},
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to new logs: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Listening for new events...")

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
