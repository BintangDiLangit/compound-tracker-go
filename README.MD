# Compound Tracker in Go

Compound Tracker is an off-chain indexing service built in Golang to track user interactions with the Compound protocol on Ethereum. The service listens for Mint and Borrow events for USDC and ETH tokens, calculates points based on token amounts and time duration, and stores the results in a PostgreSQL database.

## Features

    •	Listens for Mint and Borrow events on Compound contracts on the Ethereum Mainnet.
    •	Calculates points based on the amount of tokens and duration (with a 10-minute interval).
    •	Stores results in PostgreSQL.
    •	Provides an HTTP endpoint to retrieve the accumulated points for a specified user address.

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
  usdc: "0x39AA39c021dfbaE8faC545936693aC917d5E7563" # Contract address for USDC
  eth: "0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5" # Contract address for ETH
```

Or you can use my config (more ease to use) :

```yaml
ethereum_rpc_url: "wss://mainnet.infura.io/ws/v3/0a6fa20a5283484faf91799fc21fd148"
postgres_url: "postgres://seahorsefi:seaQWE10__@103.245.38.205/test_seahorsefi?sslmode=disable"
contracts:
  usdc: "0x39AA39c021dfbaE8faC545936693aC917d5E7563"
  eth: "0x4Ddc2D193948926D02f9B1fE9e1daa0718270ED5"
```

## Running the Program

### Clone the Repository

```
git clone https://github.com/BintangDiLangit/compound-tracker-go.git
cd compound-tracker-go
```

### Run the Program Localy

```
go run main.go
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
docker run -d -p 8082:8080 compound-tracker
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

## Notes

- This project uses Infura to listen for events on the Ethereum Mainnet.
