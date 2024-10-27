package events

import (
	"context"
	"database/sql"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/BintangDiLangit/compound-tracker/internal/config"
	"github.com/BintangDiLangit/compound-tracker/internal/db"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func HandleEvent(eventLog types.Log, client *ethclient.Client, database *sql.DB, cfg *config.Config) {

	if len(eventLog.Topics) == 0 {
		log.Println("Event log does not contain any topics, skipping this log.")
		return
	}

	eventSignature := eventLog.Topics[0].Hex()
	eventName, pointsPerUnit := getEventInfo(eventSignature, cfg)
	if eventName == "" {
		log.Println("Transaction hash for skipped event: " + eventLog.TxHash.Hex())
		return
	}

	log.Println("event name " + eventName)

	var amount *big.Int

	if len(eventLog.Data) < 32 {
		log.Println("Event data too short")
		return
	}
	/*
		The value (amount) is stored in Data because the parameter is not indexed
		and is a slice of bytes that takes the first 32 bytes of data.
		This is because in Ethereum, each parameter/value
		is stored in a 32 byte (256 bit) chunk.

		ex. for borrow event
		Data = [
			32 bytes (amount)     → Data[:32]
			32 bytes (mintTokens) → Data[32:64]
		]

		ex. for mint event
		Data = [
			32 bytes (amount)     → Data[:32]
			32 bytes (mintTokens) → Data[32:64]
		]
	*/
	amount = new(big.Int).SetBytes(eventLog.Data[:32])

	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(eventLog.BlockNumber)))
	if err != nil {
		log.Printf("Failed to fetch block timestamp: %v", err)
		return
	}

	timestamp := time.Unix(int64(block.Time()), 0)

	points := CalculatePoints(pointsPerUnit, amount, timestamp)

	db.InsertUserPoints(database, eventLog, points, eventName)
}

func getEventInfo(eventSignature string, cfg *config.Config) (string, int) {
	switch eventSignature {
	case cfg.HexMint:
		return "Mint", 1
	case cfg.HexBorrow:
		return "Borrow", 2
	default:
		return "", 0
	}
}

func CalculatePoints(pointsPerUnit int, amount *big.Int, timestamp time.Time) int64 {

	intervalDuration := 10
	duration := time.Since(timestamp)
	durationMinutes := int(duration.Minutes())

	intervals := int64(durationMinutes / intervalDuration)

	// Convert to ETH (divide by 10^18)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	ethAmount := new(big.Int).Div(amount, divisor)

	log.Println("===================")
	log.Println("Original Wei:", amount.String())
	log.Println("In ETH:", ethAmount.String())
	log.Println("Duration in minutes:", durationMinutes)
	log.Println("Intervals (10-minute):", intervals)
	log.Println("Points multiplier per unit:", pointsPerUnit)

	maxInt64 := big.NewInt(math.MaxInt64)
	if ethAmount.Cmp(maxInt64) > 0 {
		log.Println("Amount too large, skipping transaction")
		return 0
	}

	points := ethAmount.Int64()
	if points <= 0 {
		log.Println("Amount too small or negative, skipping transaction")
		return 0
	}

	finalPoints := points * intervals * int64(pointsPerUnit)
	if finalPoints < 0 {
		log.Println("Points calculation overflow, skipping transaction")
		return 0
	}

	log.Printf("Final points: %d", finalPoints)
	log.Println("===================")

	return finalPoints
}
