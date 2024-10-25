CREATE TABLE IF NOT EXISTS log_hashes (
    id SERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INTEGER NOT NULL,
    log_hash BYTEA NOT NULL
);