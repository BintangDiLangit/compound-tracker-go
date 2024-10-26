package events

import (
	"database/sql"
	"log"
	"math"
	"math/big"

	"github.com/BintangDiLangit/compound-tracker/internal/config"
	"github.com/BintangDiLangit/compound-tracker/internal/db"
	"github.com/ethereum/go-ethereum/core/types"
)

func handleEvent(eventLog types.Log, database *sql.DB, cfg *config.Config) {

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

	// Handle different event types - Transfer (Optional - For Test), Mint, Borrow
	switch eventName {
	case "Transfer":
		// Transfer just for testing because mint & borrow very rarely
		if len(eventLog.Topics) < 3 {
			log.Println("Transfer event missing amount topic")
			return
		}

		/*
			The value (amount) is stored in Topic[2] because
			- Topics[0] = event signature/hash
			- Topics[1] = from address (indexed)
			- Topics[2] = to address (indexed)
			- Topics[3] = value/amount (indexed)
		*/
		amount = new(big.Int).SetBytes(eventLog.Topics[2].Bytes())

	default:

		// For Mint and Borrow events
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
	}

	points := calculatePoints(pointsPerUnit, amount)

	db.InsertUserPoints(database, eventLog, points, eventName)
}

func getEventInfo(eventSignature string, cfg *config.Config) (string, int) {
	switch eventSignature {
	case cfg.HexMint:
		return "Mint", 1
	case cfg.HexBorrow:
		return "Borrow", 2
	// case cfg.HexTransfer: // transfer just for testing
	// 	return "Transfer", 0
	default:
		return "", 0
	}
}

func calculatePoints(pointsPerUnit int, amount *big.Int) int64 {
	// Convert to ETH first (divide by 10^18)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	ethAmount := new(big.Int).Div(amount, divisor)

	log.Println("===================")
	log.Println("Original Wei:", amount.String())
	log.Println("In ETH:", ethAmount.String())
	log.Println("Points multiplier:", pointsPerUnit)

	// Check if amount would overflow int64
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

	finalPoints := points * int64(pointsPerUnit)
	if finalPoints < 0 {
		log.Println("Points calculation overflow, skipping transaction")
		return 0
	}

	log.Printf("Final points: %d", finalPoints)
	log.Println("===================")

	return finalPoints
}
