CREATE TABLE IF NOT EXISTS log_hashes (
    id SERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL,
    transaction_hash TEXT NOT NULL,
    log_index INTEGER NOT NULL,
    log_hash BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(block_number, transaction_hash, log_index)
);