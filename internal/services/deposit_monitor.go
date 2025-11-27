package services

import (
	"context"
	"fmt"
	"time"

	"github.com/dechat/exchange-service/internal/adapters"
	"github.com/dechat/exchange-service/internal/models"
	"github.com/dechat/exchange-service/internal/storage"
)

type DepositMonitor struct {
	storage  storage.Storage
	adapters map[models.Chain]adapters.BlockchainAdapter
	configs  map[models.Chain]ChainMonitorConfig
}

type ChainMonitorConfig struct {
	MinConfirmations int
	PollInterval     time.Duration
}

func NewDepositMonitor(storage storage.Storage, adapters map[models.Chain]adapters.BlockchainAdapter) *DepositMonitor {
	configs := make(map[models.Chain]ChainMonitorConfig)
	// Default configs
	for chain := range adapters {
		configs[chain] = ChainMonitorConfig{
			MinConfirmations: 1,
			PollInterval:     5 * time.Second,
		}
	}

	return &DepositMonitor{
		storage:  storage,
		adapters: adapters,
		configs:  configs,
	}
}

// Start starts monitoring deposits for all chains
func (m *DepositMonitor) Start(ctx context.Context) error {
	for chain := range m.adapters {
		go m.monitorChain(ctx, chain)
	}
	return nil
}

func (m *DepositMonitor) monitorChain(ctx context.Context, chain models.Chain) {
	config := m.configs[chain]
	ticker := time.NewTicker(config.PollInterval)
	defer ticker.Stop()

	var lastBlock int64 = 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get latest block
			adapter := m.adapters[chain]
			latestBlock, err := adapter.GetLatestBlock(ctx)
			if err != nil {
				fmt.Printf("Error getting latest block for %s: %v\n", chain, err)
				continue
			}

			// Process blocks from lastBlock+1 to latestBlock
			for blockNum := lastBlock + 1; blockNum <= latestBlock; blockNum++ {
				if err := m.processBlock(ctx, chain, blockNum, config.MinConfirmations); err != nil {
					fmt.Printf("Error processing block %d for %s: %v\n", blockNum, chain, err)
				}
			}

			lastBlock = latestBlock
		}
	}
}

func (m *DepositMonitor) processBlock(ctx context.Context, chain models.Chain, blockNumber int64, minConfirmations int) error {
	adapter := m.adapters[chain]
	transactions, err := adapter.GetBlockTransactions(ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to get block transactions: %w", err)
	}

	for _, tx := range transactions {
		// Check if this is a deposit to one of our addresses
		deposit, err := m.storage.GetDepositByAddress(chain, tx.To)
		if err != nil {
			return fmt.Errorf("failed to get deposit: %w", err)
		}

		if deposit == nil {
			continue // Not our address
		}

		// Check if already processed
		if deposit.TxHash == tx.Hash {
			continue
		}

		// Get transaction status
		status, err := adapter.GetTransactionStatus(ctx, tx.Hash)
		if err != nil {
			return fmt.Errorf("failed to get tx status: %w", err)
		}

		// Update deposit
		deposit.TxHash = tx.Hash
		deposit.ReceivedAmount = tx.Amount.String()
		deposit.BlockNumber = status.BlockNumber
		deposit.Confirmations = status.Confirmations

		if status.Confirmations >= minConfirmations {
			deposit.Status = models.DepositStatusConfirmed
			now := time.Now()
			deposit.ConfirmedAt = &now
		}

		if err := m.storage.UpdateDeposit(deposit); err != nil {
			return fmt.Errorf("failed to update deposit: %w", err)
		}

		// TODO: Send webhook/event about deposit confirmation
		fmt.Printf("Deposit confirmed: chain=%s, order_id=%s, amount=%s\n", chain, deposit.OrderID, deposit.ReceivedAmount)
	}

	return nil
}

