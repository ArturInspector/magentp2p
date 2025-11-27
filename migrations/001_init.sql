-- Create chains enum
CREATE TYPE chain_type AS ENUM ('ethereum', 'polygon', 'bsc', 'arbitrum', 'optimism');
CREATE TYPE deposit_status_type AS ENUM ('pending', 'confirmed', 'expired');
CREATE TYPE withdrawal_status_type AS ENUM ('pending', 'sent', 'confirmed', 'failed');

-- Deposits table
CREATE TABLE deposits (
    id BIGSERIAL PRIMARY KEY,
    chain chain_type NOT NULL,
    address VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    order_id VARCHAR(255) NOT NULL,
    expected_amount VARCHAR(255) NOT NULL,
    received_amount VARCHAR(255) DEFAULT '0',
    tx_hash VARCHAR(255),
    block_number BIGINT,
    confirmations INTEGER DEFAULT 0,
    status deposit_status_type DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    confirmed_at TIMESTAMP
);

CREATE INDEX idx_deposits_address ON deposits(address);
CREATE INDEX idx_deposits_order_id ON deposits(order_id);
CREATE INDEX idx_deposits_tx_hash ON deposits(tx_hash);
CREATE INDEX idx_deposits_status ON deposits(status);

-- Withdrawals table
CREATE TABLE withdrawals (
    id BIGSERIAL PRIMARY KEY,
    chain chain_type NOT NULL,
    order_id VARCHAR(255) NOT NULL,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount VARCHAR(255) NOT NULL,
    fee VARCHAR(255) NOT NULL,
    tx_hash VARCHAR(255),
    status withdrawal_status_type DEFAULT 'pending',
    block_number BIGINT,
    confirmations INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    sent_at TIMESTAMP,
    confirmed_at TIMESTAMP
);

CREATE INDEX idx_withdrawals_order_id ON withdrawals(order_id);
CREATE INDEX idx_withdrawals_tx_hash ON withdrawals(tx_hash);
CREATE INDEX idx_withdrawals_status ON withdrawals(status);

-- Hot wallets table
CREATE TABLE hot_wallets (
    id BIGSERIAL PRIMARY KEY,
    chain chain_type NOT NULL UNIQUE,
    address VARCHAR(255) NOT NULL,
    encrypted_key TEXT NOT NULL,
    balance VARCHAR(255) DEFAULT '0',
    last_checked_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_hot_wallets_chain ON hot_wallets(chain);

-- Transactions table
CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    chain chain_type NOT NULL,
    tx_hash VARCHAR(255) NOT NULL UNIQUE,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    amount VARCHAR(255) NOT NULL,
    fee VARCHAR(255) NOT NULL,
    block_number BIGINT,
    status VARCHAR(50) DEFAULT 'pending',
    confirmations INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX idx_transactions_chain ON transactions(chain);
CREATE INDEX idx_transactions_to_address ON transactions(to_address);

