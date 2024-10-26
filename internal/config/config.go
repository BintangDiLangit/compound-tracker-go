package config

import (
	"log"
	"math/big"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	EthereumRPCURL string `yaml:"ethereum_rpc_url"`
	PostgresURL    string `yaml:"postgres_url"`
	Contracts      struct {
		USDC string `yaml:"usdc"`
		ETH  string `yaml:"eth"`
	} `yaml:"contracts"`
	HexMint     string `yaml:"hex_mint"`
	HexBorrow   string `yaml:"hex_borrow"`
	HexTransfer string `yaml:"hex_transfer"`
}

type EventDetails struct {
	Name          string
	PointsPerUnit *big.Int
	Amount        *big.Int
	Address       string
	BlockNumber   uint64
	TxHash        string
	LogIndex      uint
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	log.Println("Configuration loaded successfully.")
	return &cfg, nil
}
