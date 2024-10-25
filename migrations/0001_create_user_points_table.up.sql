CREATE TABLE IF NOT EXISTS user_points (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL,
    points INTEGER NOT NULL,
    event VARCHAR(10) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    block_number BIGINT NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INTEGER,
    amount DECIMAL(38,18) NOT NULL,
    token_type VARCHAR(10)
);