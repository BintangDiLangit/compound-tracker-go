package config

import (
	"log"
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
	HexMint   string `yaml:"hex_mint"`
	HexBorrow string `yaml:"hex_borrow"`
}

func LoadConfig(configFile string) (*Config, error) {
	var cfg Config
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
		return nil, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
		return nil, err
	}
	return &cfg, nil
}
