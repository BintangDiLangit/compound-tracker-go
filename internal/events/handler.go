package events

import (
	"context"
	"database/sql"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"

	"compound-tracker/internal/config"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func ListenForEvents(client *ethclient.Client, db *sql.DB, cfg *config.Config) {
	usdcAddress := common.HexToAddress(cfg.Contracts.USDC)
	ethAddress := common.HexToAddress(cfg.Contracts.ETH)

	// Specific Topic
	fromBlock := getLastProcessedBlock()

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(fromBlock),
		ToBlock:   nil,
		Addresses: []common.Address{usdcAddress, ethAddress},
		Topics: [][]common.Hash{
			{
				common.HexToHash(cfg.HexMint),   // Mint
				common.HexToHash(cfg.HexBorrow), // Borrow
			},
		},
	}

	// All Topic for Test

	/*
		latestBlockNumber, err := client.BlockNumber(context.Background())
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(latestBlockNumber) - 10000),
			ToBlock:   nil, // Until the latest block
		}
		if err != nil {
			log.Fatalf("Failed to fetch block number: %v", err)
		}
		log.Printf("Current block number: %d", latestBlockNumber)

		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(latestBlockNumber) - 10000),
			ToBlock:   nil,
			Addresses: []common.Address{usdcAddress, ethAddress},
		}
	*/

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Starting to listen for events...")
	for {
		select {

		case err := <-sub.Err():
			log.Printf("Error: %v", err)
		case vLog := <-logs:
			log.Printf("Received log: %+v", vLog)
			handleEvent(vLog, db, cfg)

			blockNumber := vLog.BlockNumber
			saveLastProcessedBlock(int64(blockNumber))
		}

	}
}

func handleEvent(eventLog types.Log, db *sql.DB, cfg *config.Config) {
	log.Printf("Event received with signature: %s", eventLog.Topics[0].Hex())

	eventName := ""
	pointsPerUnit := 0

	switch eventLog.Topics[0].Hex() {
	case cfg.HexMint:
		eventName = "Mint"
		pointsPerUnit = 1
	case cfg.HexBorrow:
		eventName = "Borrow"
		pointsPerUnit = 2
	default:
		eventName = "Others"
		pointsPerUnit = 3
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

func getLastProcessedBlock() int64 {
	if _, err := os.Stat("last_block.txt"); os.IsNotExist(err) {
		log.Println("No last_block.txt found, starting from latest block.")
		return 0
	}

	data, err := ioutil.ReadFile("last_block.txt")
	if err != nil {
		log.Printf("Failed to read last processed block, starting from latest: %v", err)
		return 0
	}

	blockNum, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		log.Printf("Failed to parse last block number, starting from latest: %v", err)
		return 0
	}
	return blockNum
}

func saveLastProcessedBlock(blockNumber int64) {
	err := ioutil.WriteFile("last_block.txt", []byte(strconv.FormatInt(blockNumber, 10)), 0644)
	if err != nil {
		log.Printf("Failed to save last processed block: %v", err)
	}
}
