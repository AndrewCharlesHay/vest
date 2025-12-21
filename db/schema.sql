CREATE TABLE IF NOT EXISTS positions (
    date DATE NOT NULL,
    account_id VARCHAR(50) NOT NULL,
    ticker VARCHAR(50) NOT NULL,
    quantity NUMERIC(18, 4),
    market_value NUMERIC(18, 2),
    shares NUMERIC(18, 4), -- from Format 2
    source_system VARCHAR(50),
    ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (date, account_id, ticker)
);

-- Index for efficient querying by date and account
CREATE INDEX idx_positions_date_account ON positions (date, account_id);
