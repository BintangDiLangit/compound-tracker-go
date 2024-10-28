package events

import (
	"context"
	"database/sql"
	"log"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/BintangDiLangit/compound-tracker/internal/config"
	"github.com/BintangDiLangit/compound-tracker/internal/db"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var contractABI abi.ABI

type MintEvent struct {
	Account     common.Address
	Amount      *big.Int `abi:"amount"`
	TotalSupply *big.Int `abi:"totalSupply"`
}

type BorrowEvent struct {
	Account     common.Address
	Amount      *big.Int `abi:"amount"`
	TotalSupply *big.Int `abi:"totalSupply"`
}

func init() {
	const contractABIString = `[
        {
			"anonymous": false,
			"inputs": [
				{
				"indexed": true,
				"name": "account",
				"type": "address"
				},
				{
				"indexed": false,
				"name": "amount",
				"type": "uint256"
				},
				{
				"indexed": false,
				"name": "totalSupply",
				"type": "uint256"
				}
			],
			"name": "Mint",
			"type": "event"
        },
		{
			"anonymous": false,
			"inputs": [
				{
				"indexed": true,
				"name": "account",
				"type": "address"
				},
				{
				"indexed": false,
				"name": "amount",
				"type": "uint256"
				},
				{
				"indexed": false,
				"name": "totalSupply",
				"type": "uint256"
				}
			],
			"name": "Borrow",
			"type": "event"
		}
    ]`

	var err error
	contractABI, err = abi.JSON(strings.NewReader(contractABIString))
	if err != nil {
		log.Fatal(err)
	}
}

func HandleEvent(eventLog types.Log, client *ethclient.Client, database *sql.DB, cfg *config.Config) {
	if len(eventLog.Topics) == 0 {
		// log.Println("Event log does not contain any topics, skipping this log.")
		return
	}

	eventSignature := eventLog.Topics[0].Hex()
	eventName, pointsPerUnit := getEventInfo(eventSignature, cfg)
	if eventName == "" {
		// log.Println("Transaction hash for skipped event: " + eventLog.TxHash.Hex())
		return
	}

	log.Println("Event Name " + eventName)

	// Unpack event data
	var weiAmount *big.Int
	if eventName == "Borrow" {
		event := new(BorrowEvent)
		err := contractABI.UnpackIntoInterface(event, "Borrow", eventLog.Data)
		if err != nil {
			log.Printf("Failed to unpack event: %v", err)
			// Fallback method if fail
			if len(eventLog.Data) >= 32 {
				weiAmount = new(big.Int).SetBytes(eventLog.Data[:32])
			}
		} else {
			weiAmount = event.Amount
		}
	} else if eventName == "Mint" {
		event := new(MintEvent)
		err := contractABI.UnpackIntoInterface(event, "Mint", eventLog.Data)
		if err != nil {
			log.Printf("Failed to unpack event: %v", err)
			if len(eventLog.Data) >= 32 {
				weiAmount = new(big.Int).SetBytes(eventLog.Data[:32])
			}
		} else {
			weiAmount = event.Amount
		}
	}

	if weiAmount == nil || weiAmount.Sign() == 0 {
		log.Printf("Invalid or zero amount for event %s", eventName)
		return
	}

	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(eventLog.BlockNumber)))
	if err != nil {
		log.Printf("Failed to fetch block timestamp: %v", err)
		return
	}

	timestamp := time.Unix(int64(block.Time()), 0)

	intervalDuration := 10
	duration := time.Since(timestamp)
	durationMinutes := int(duration.Minutes())

	intervals := int64(durationMinutes / intervalDuration)

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	ethAmount := new(big.Int).Div(weiAmount, divisor)

	points := CalculatePoints(pointsPerUnit, weiAmount, ethAmount, timestamp, durationMinutes, int64(intervals),
		duration)

	db.InsertUserPoints(database, eventLog, points, eventName, weiAmount.String(), ethAmount.String(), durationMinutes)
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

func CalculatePoints(pointsPerUnit int, weiAmount *big.Int, ethAmount *big.Int,
	timestamp time.Time, durationMinutes int, intervals int64, duration time.Duration) int64 {

	log.Println("===================")
	// Reference time in go
	log.Printf("Transaction Time: %v", timestamp.Format("2006-01-02 15:04:05"))
	log.Printf("Current Time: %v", time.Now().Format("2006-01-02 15:04:05"))

	log.Printf("Time Difference: %v", duration.Round(time.Second))
	log.Println("Original Wei:", weiAmount.String())
	log.Println("In ETH:", ethAmount.String())
	log.Println("Duration in minutes:", durationMinutes)
	log.Println("Intervals (10-minute):", intervals)
	log.Println("Points multiplier per unit:", pointsPerUnit)
	log.Println("===================")

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
