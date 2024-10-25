package events

import (
	"context"
	"database/sql"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"compound-tracker/internal/config"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
				common.HexToHash(cfg.HexMint),
				common.HexToHash(cfg.HexBorrow),
			},
		},
	}

	// All Topic for Test
	/* latestBlockNumber, err := client.BlockNumber(context.Background())
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(latestBlockNumber) - 10000),
		ToBlock:   nil, // Until the latest block
	}
	if err != nil {
		log.Fatalf("Failed to fetch block number: %v", err)
	}
	log.Printf("Current block number: %d", latestBlockNumber)
	*/

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to logs: %v", err)
	}
	defer sub.Unsubscribe()

	log.Println("Starting to listen for events...")
	for {
		// Create a new subscription
		sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
		if err != nil {
			log.Printf("Failed to subscribe to logs: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for {
			select {
			case err := <-sub.Err():
				log.Printf("Subscription error: %v. Reconnecting...", err)
				time.Sleep(5 * time.Second)
				continue
			case vLog := <-logs:
				log.Printf("Received log: %+v", vLog)
				handleEvent(vLog, db, cfg)

				blockNumber := vLog.BlockNumber
				saveLastProcessedBlock(int64(blockNumber))
			}
		}
	}
}

func handleEvent(eventLog types.Log, db *sql.DB, cfg *config.Config) {
	log.Printf("Event received with signature: %s", eventLog.Topics[0].Hex())

	eventSignature := eventLog.Topics[0].Hex()

	var eventName string
	var pointsPerUnit int
	var tokenType string

	if len(eventLog.Data) < 32 {
		log.Println("Log data is too short, skipping this log")
		return
	}

	switch eventSignature {
	case cfg.HexMint:
		eventName = "Mint"
		pointsPerUnit = 1
		tokenType = "USDC"
	case cfg.HexBorrow:
		eventName = "Borrow"
		pointsPerUnit = 2
		tokenType = "ETH"
	default:
		log.Println("Unknown event type, skipping")
		return
	}

	amount := new(big.Int)
	amount.SetBytes(eventLog.Data[:32])

	duration := 10
	points := pointsPerUnit * int(amount.Int64()) * (duration / 10)

	log.Printf("Event: %s, Amount: %s, Points: %d", eventName, amount.String(), points)

	_, err := db.Exec(`
	INSERT INTO user_points 
	(address, points, event, block_number, transaction_hash, log_index, amount, token_type) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		eventLog.Address.Hex(),
		points,
		eventName,
		eventLog.BlockNumber,
		eventLog.TxHash.Hex(),
		eventLog.Index,
		amount.String(),
		tokenType)
	if err != nil {
		log.Printf("Failed to insert points: %v", err)
	}

	logHash := hashLog(eventLog)
	storeLogHashInDB(db, eventLog.BlockNumber, eventLog.TxHash.Hex(), eventLog.Index, logHash)

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

func storeLogHashInDB(db *sql.DB, blockNumber uint64, txHash string, logIndex uint, logHash []byte) {
	_, err := db.Exec(`
        INSERT INTO log_hashes (block_number, transaction_hash, log_index, log_hash) 
        VALUES ($1, $2, $3, $4)`,
		blockNumber, txHash, logIndex, logHash)
	if err != nil {
		log.Printf("Failed to insert log hash: %v", err)
	}
}

func hashLog(eventLog types.Log) []byte {
	return crypto.Keccak256(eventLog.Data)
}
