CREATE TABLE IF NOT EXISTS user_points (
    id SERIAL PRIMARY KEY,
    address TEXT NOT NULL,
    points BIGINT NOT NULL,
    event TEXT NOT NULL,
    block_number BIGINT NOT NULL,
    transaction_hash TEXT NOT NULL,
    log_index INTEGER NOT NULL,
    amount TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(transaction_hash, log_index)
);