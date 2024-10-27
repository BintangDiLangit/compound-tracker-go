package integrations

import (
	"database/sql"
	"math/big"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/BintangDiLangit/compound-tracker/internal/config"
	"github.com/BintangDiLangit/compound-tracker/internal/events"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}
func TestHandleEvent(t *testing.T) {
	// Setup test config
	cfg, err := config.Load(filepath.Join(getProjectRoot(), "config.yaml"))
	require.NoError(t, err, "Should load config")

	// Setup database
	database, err := sql.Open("postgres", cfg.PostgresURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Test database connection
	if err := database.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Setup Ethereum client
	if cfg.EthereumRPCURL == "" {
		t.Skip("ETH_RPC_URL not set")
	}
	client, err := ethclient.Dial(cfg.EthereumRPCURL)
	if err != nil {
		t.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name    string
		log     types.Log
		wantErr bool
	}{
		{
			name: "Valid Mint Event",
			log: types.Log{
				Address: common.HexToAddress("0x4ddc2d193948926d02f9b1fe9e1daa0718270ed5"),
				Topics: []common.Hash{
					common.HexToHash(cfg.HexMint),
					common.HexToHash("0x000000000000000000000000" + "1234567890123456789012345678901234567890"),
				},
				Data:        makeTestData(big.NewInt(1e18)), // 1 ETH
				BlockNumber: 12345678,
				TxHash:      common.HexToHash("0x1234"),
				Index:       0,
			},
		},
		{
			name: "Valid Borrow Event",
			log: types.Log{
				Address: common.HexToAddress("0x4ddc2d193948926d02f9b1fe9e1daa0718270ed5"),
				Topics: []common.Hash{
					common.HexToHash(cfg.HexBorrow),
					common.HexToHash("0x000000000000000000000000" + "1234567890123456789012345678901234567890"),
				},
				Data:        makeTestData(big.NewInt(2e18)), // 2 ETH
				BlockNumber: 12345678,
				TxHash:      common.HexToHash("0x5678"),
				Index:       0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up previous test data
			_, err := database.Exec("DELETE FROM user_points")
			require.NoError(t, err)

			events.HandleEvent(tt.log, client, database, cfg)

			if !tt.wantErr {
				var count int
				err := database.QueryRow(`
                    SELECT COUNT(*) 
                    FROM user_points 
                    WHERE transaction_hash = $1
                `, tt.log.TxHash.Hex()).Scan(&count)

				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func makeTestData(amount *big.Int) []byte {
	data := make([]byte, 32)
	amount.FillBytes(data)
	return data
}

func TestCalculatePoints(t *testing.T) {
	tests := []struct {
		name          string
		pointsPerUnit int
		amount        *big.Int
		timestamp     time.Time
		want          int64
	}{
		{
			name:          "Mint Points",
			pointsPerUnit: 1,
			amount:        big.NewInt(1e18), // 1 ETH
			timestamp:     time.Now().Add(-20 * time.Minute),
			want:          2,
		},
		{
			name:          "Borrow Points",
			pointsPerUnit: 2,
			amount:        big.NewInt(1e18), // 1 ETH
			timestamp:     time.Now().Add(-30 * time.Minute),
			want:          6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := events.CalculatePoints(tt.pointsPerUnit, tt.amount, tt.timestamp)
			assert.Equal(t, tt.want, got)
		})
	}
}
