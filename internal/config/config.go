package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	EthereumRPCURL string `yaml:"ethereum_rpc_url"`
	PostgresURL    string `yaml:"postgres_url"`
	Contracts      struct {
		ETH string `yaml:"eth"`
	} `yaml:"contracts"`
	HexMint   string `yaml:"hex_mint"`
	HexBorrow string `yaml:"hex_borrow"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
