package utils

import (
	"log"
	"os"
	"strconv"
)

func GetLastProcessedBlock() int64 {
	if _, err := os.Stat("last_block.txt"); os.IsNotExist(err) {
		return 0
	}

	data, err := os.ReadFile("last_block.txt")
	if err != nil {
		log.Printf("Failed to read last processed block: %v", err)
		return 0
	}

	// Handle empty file
	if len(data) == 0 {
		log.Println("last_block.txt is empty, starting from latest block.")
		return 0
	}

	blockNum, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		log.Printf("Failed to parse last block number: %v", err)
		return 0
	}
	return blockNum
}

func SaveLastProcessedBlock(blockNumber int64) {
	err := os.WriteFile("last_block.txt", []byte(strconv.FormatInt(blockNumber, 10)), 0644)
	if err != nil {
		log.Printf("Failed to save last processed block: %v", err)
	}
}
