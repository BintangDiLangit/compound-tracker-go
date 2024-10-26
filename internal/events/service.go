package events

import (
	"database/sql"

	"github.com/BintangDiLangit/compound-tracker/internal/config"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Service struct {
	client *ethclient.Client
	db     *sql.DB
	cfg    *config.Config
}

func NewService(client *ethclient.Client, db *sql.DB, cfg *config.Config) *Service {
	return &Service{
		client: client,
		db:     db,
		cfg:    cfg,
	}
}

func (s *Service) StartListening() {
	ListenForEvents(s.client, s.db, s.cfg)
}
