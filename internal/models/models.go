package models

import "time"

// Chain represents supported blockchain
type Chain string

const (
	ChainEthereum Chain = "ethereum"
	ChainPolygon  Chain = "polygon"
	ChainBSC      Chain = "bsc"
	ChainArbitrum Chain = "arbitrum"
	ChainOptimism Chain = "optimism"
)

// DepositStatus represents deposit status
type DepositStatus string

const (
	DepositStatusPending   DepositStatus = "pending"
	DepositStatusConfirmed DepositStatus = "confirmed"
	DepositStatusExpired   DepositStatus = "expired"
)

// WithdrawalStatus represents withdrawal status
type WithdrawalStatus string

const (
	WithdrawalStatusPending   WithdrawalStatus = "pending"
	WithdrawalStatusSent      WithdrawalStatus = "sent"
	WithdrawalStatusConfirmed WithdrawalStatus = "confirmed"
	WithdrawalStatusFailed    WithdrawalStatus = "failed"
)

// Deposit represents user deposit
type Deposit struct {
	ID             int64         `db:"id" json:"id"`
	Chain          Chain         `db:"chain" json:"chain"`
	Address        string        `db:"address" json:"address"`
	UserID         string        `db:"user_id" json:"user_id"`
	OrderID        string        `db:"order_id" json:"order_id"`
	ExpectedAmount string        `db:"expected_amount" json:"expected_amount"`
	ReceivedAmount string        `db:"received_amount" json:"received_amount"`
	TxHash         string        `db:"tx_hash" json:"tx_hash"`
	BlockNumber    int64         `db:"block_number" json:"block_number"`
	Confirmations  int           `db:"confirmations" json:"confirmations"`
	Status         DepositStatus `db:"status" json:"status"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
	ConfirmedAt    *time.Time    `db:"confirmed_at" json:"confirmed_at"`
}

// Withdrawal represents user withdrawal
type Withdrawal struct {
	ID            int64            `db:"id" json:"id"`
	Chain         Chain            `db:"chain" json:"chain"`
	OrderID       string           `db:"order_id" json:"order_id"`
	FromAddress   string           `db:"from_address" json:"from_address"`
	ToAddress     string           `db:"to_address" json:"to_address"`
	Amount        string           `db:"amount" json:"amount"`
	Fee           string           `db:"fee" json:"fee"`
	TxHash        string           `db:"tx_hash" json:"tx_hash"`
	Status        WithdrawalStatus `db:"status" json:"status"`
	BlockNumber   int64            `db:"block_number" json:"block_number"`
	Confirmations int              `db:"confirmations" json:"confirmations"`
	CreatedAt     time.Time        `db:"created_at" json:"created_at"`
	SentAt        *time.Time        `db:"sent_at" json:"sent_at"`
	ConfirmedAt   *time.Time        `db:"confirmed_at" json:"confirmed_at"`
}

// HotWallet represents hot wallet for a chain
type HotWallet struct {
	ID                int64     `db:"id" json:"id"`
	Chain             Chain     `db:"chain" json:"chain"`
	Address           string    `db:"address" json:"address"`
	EncryptedKey      string    `db:"encrypted_key" json:"-"` // не возвращаем в API
	Balance           string    `db:"balance" json:"balance"`
	LastCheckedAt     time.Time `db:"last_checked_at" json:"last_checked_at"`
}

// Transaction represents blockchain transaction
type Transaction struct {
	ID            int64     `db:"id" json:"id"`
	Chain         Chain     `db:"chain" json:"chain"`
	TxHash        string    `db:"tx_hash" json:"tx_hash"`
	FromAddress   string    `db:"from_address" json:"from_address"`
	ToAddress     string    `db:"to_address" json:"to_address"`
	Amount        string    `db:"amount" json:"amount"`
	Fee           string    `db:"fee" json:"fee"`
	BlockNumber   int64     `db:"block_number" json:"block_number"`
	Status        string    `db:"status" json:"status"`
	Confirmations int       `db:"confirmations" json:"confirmations"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

