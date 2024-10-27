# Compound Tracker in Go

Compound Tracker is an off-chain indexing service built in Golang to track user interactions with the Compound protocol on Ethereum. The service listens for Mint and Borrow events for ETH tokens, calculates points based on token amounts and time duration, and stores the results in a PostgreSQL database.

## Features

    •	Listens for Mint and Borrow events on Compound contracts on the Ethereum Mainnet.
    •	Calculates points based on the amount of tokens and duration (with a 10-minute interval).
    •	Stores results in PostgreSQL.
    •	Provides an HTTP endpoint to retrieve the accumulated points for a specified user address.
    •	Live URL : https://compound-tracker.bintangmfhd.com/points?address=[yourAddress]

## Configuration

### Prerequisites

    •	Go 1.22 or higher
    •	PostgreSQL
    •	Infura or any other Ethereum node provider with WebSocket support

### Configuration File

Create a config.yaml file in the project root with the following format:

```yaml
ethereum_rpc_url: "wss://mainnet.infura.io/ws/v3/YOUR_INFURA_PROJECT_ID"
postgres_url: "postgres://USER:PASSWORD@HOST/DATABASE?sslmode=disable"
contracts:
  eth: "0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5" # Contract address for ETH
hex_mint: " 0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f"
hex_borrow: "0x13ed6866d4e1ee6da46f845c46d7e54120883d75c5ea9a2dacc1c4ca8984ab80"
hex_transfer: "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
```

Or you can use my config (more ease to use) :

```yaml
ethereum_rpc_url: "wss://mainnet.infura.io/ws/v3/0a6fa20a5283484faf91799fc21fd148"
postgres_url: "postgres://seahorsefi:seaQWE10__@103.245.38.205/test_seahorsefi?sslmode=disable"
contracts:
  eth: "0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5"
hex_mint: " 0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f"
hex_borrow: "0x13ed6866d4e1ee6da46f845c46d7e54120883d75c5ea9a2dacc1c4ca8984ab80"
hex_transfer: "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
```

## Running the Program

### Clone the Repository

```
git clone https://github.com/BintangDiLangit/compound-tracker-go.git
cd compound-tracker-go
```

### Run the Program Localy

```
go run cmd/compound-tracker/main.go
```

#### Testing

```
go test -v ./test/integrations/...
```

The program will connect to Ethereum and PostgreSQL, perform database migrations, and start listening for events.

### HTTP Endpoint:

    1. GET /points: Retrieve the total points for a given Ethereum address.
    2. Example usage with curl:

```
curl "http://localhost:8082/points?address=0xYourEthereumAddress"
```

### Or Run with Docker

```
docker build -t compound-tracker .
docker run -d -p 8082:8082 compound-tracker
```

## Testing

If there is no active Mint or Borrow event on the Ethereum network, you can add a mock event manually in main.go for testing purposes (refer to the example in the code). Be sure to remove the mock event after testing.

## Database Structure

The user_points table stores calculated points based on Mint or Borrow events.

```sql
CREATE TABLE user_points (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL,
    points INTEGER NOT NULL,
    event VARCHAR(10) NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Bonus Point Handling

This section outlines how the implemented system addresses the bonus challenges provided in the task:

### 1. What if the Server Goes Down? How Should We Replay the Transactions?

Solution:

- The system is designed to persist the last processed block number in a file (last_block.txt) after each successful log processing. This allows the system to replay any missed transactions after the server goes down.
- Upon restarting, the system reads the last processed block from last_block.txt and continues processing logs starting from that block. This ensures no events are missed, even during server downtime.

### 2. What if the Node/Connection is Down?

Solution:

- The system incorporates an automatic reconnection mechanism in case the connection to the Ethereum node (e.g., Infura) goes down.
- If the node or connection drops, the system logs the error and then waits for 5 seconds before retrying the connection. This retry mechanism ensures that the connection is re-established and event processing resumes automatically without manual intervention

## 3. How to Maximize Uptime to Listen to Events?

Solution:
The system is built to maximize uptime through the following strategies:

- Automatic Reconnection: As described above, if the Ethereum node or network connection goes down, the system automatically retries to reconnect without manual intervention.
- Last Processed Block Persistence: By saving the last processed block, the system ensures that no logs are missed even during downtime. The logs are processed starting from the last saved block upon reconnection.
- Multiple Instances for Redundancy (optional): For further resilience, you can deploy multiple instances of the service across different servers or regions. This allows for high availability, as one instance can take over if another goes down.

## 4. Is There a Way We Can Prove We Indexed the Points Honestly? [Brevis Co-Processor]

Solution:

- To ensure transparency and verifiability, you can leverage Merkle Trees to cryptographically prove that the events have been indexed honestly.
- For each log processed, a hash of the log can be stored in a Merkle Tree. At regular intervals (e.g., daily), the Merkle Root can be published to an external verifier or even stored on-chain to provide a tamper-proof record of the processed logs.

## Notes

- This project uses Infura to listen for events on the Ethereum Mainnet.
