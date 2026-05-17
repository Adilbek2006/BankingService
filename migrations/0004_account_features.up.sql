ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS tier VARCHAR(20) DEFAULT 'BASIC',
    ADD COLUMN IF NOT EXISTS daily_limit DECIMAL(15, 2) DEFAULT 0.0,
    ADD COLUMN IF NOT EXISTS monthly_limit DECIMAL(15, 2) DEFAULT 0.0;

CREATE TABLE IF NOT EXISTS cards (
    card_id VARCHAR(50) PRIMARY KEY,
    account_id VARCHAR(50) NOT NULL,
    card_type VARCHAR(20) NOT NULL,
    number VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    limit_amount DECIMAL(15, 2) DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_cards_account FOREIGN KEY (account_id) REFERENCES accounts(account_id)
);

CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards (account_id);
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts (user_id);

