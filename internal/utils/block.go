package utils

import (
	"encoding/json"
	"os"
)

const stateFile = "last_block.json"

type State struct {
	LastProcessedBlock int64 `json:"last_processed_block"`
}

func SaveLastProcessedBlock(blockNumber int64) error {
	state := State{LastProcessedBlock: blockNumber}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0644)
}

func GetLastProcessedBlock() int64 {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return 0
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return 0
	}
	return state.LastProcessedBlock
}
