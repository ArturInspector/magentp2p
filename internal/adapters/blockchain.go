package adapters

import (
	"context"
	"math/big"
)

// BlockchainAdapter interface for different blockchains
type BlockchainAdapter interface {
	// GenerateAddress generates new deposit address
	GenerateAddress(ctx context.Context) (string, error)

	// GetBalance returns balance of address
	GetBalance(ctx context.Context, address string) (*big.Int, error)

	// SendTransaction sends transaction and returns tx hash
	SendTransaction(ctx context.Context, fromAddress, toAddress string, amount *big.Int, privateKey string) (string, error)

	// GetTransactionStatus returns transaction status and confirmations
	GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error)

	// GetLatestBlock returns latest block number
	GetLatestBlock(ctx context.Context) (int64, error)

	// GetBlockTransactions returns transactions from block
	GetBlockTransactions(ctx context.Context, blockNumber int64) ([]*Transaction, error)

	// GetGasPrice returns current gas price
	GetGasPrice(ctx context.Context) (*big.Int, error)
}

type TransactionStatus struct {
	Status       string
	BlockNumber  int64
	Confirmations int
	Success      bool
}

type Transaction struct {
	Hash     string
	From     string
	To       string
	Amount   *big.Int
	BlockNum int64
}

