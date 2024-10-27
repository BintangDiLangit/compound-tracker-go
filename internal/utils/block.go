package utils

import (
	"log"
	"os"
	"strconv"
)

func SaveLastProcessedBlock(blockNumber int64) {
	err := os.WriteFile("last_block.txt", []byte(strconv.FormatInt(blockNumber, 10)), 0644)
	if err != nil {
		log.Printf("Failed to save last processed block: %v", err)
	}
}
