package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/dechat/exchange-service/internal/adapters"
	"github.com/dechat/exchange-service/internal/models"
	"github.com/dechat/exchange-service/internal/storage"
)

type WalletService struct {
	storage storage.Storage
	adapters map[models.Chain]adapters.BlockchainAdapter
}

func NewWalletService(storage storage.Storage, adapters map[models.Chain]adapters.BlockchainAdapter) *WalletService {
	return &WalletService{
		storage: storage,
		adapters: adapters,
	}
}

// GenerateDepositAddress generates new deposit address for user
func (s *WalletService) GenerateDepositAddress(ctx context.Context, chain models.Chain, userID, orderID string) (string, error) {
	adapter, ok := s.adapters[chain]
	if !ok {
		return "", fmt.Errorf("chain %s not supported", chain)
	}

	address, err := adapter.GenerateAddress(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to generate address: %w", err)
	}

	// Save deposit record
	deposit := &models.Deposit{
		Chain:          chain,
		Address:        address,
		UserID:         userID,
		OrderID:        orderID,
		ExpectedAmount: "0", // TODO: get from order
		Status:         models.DepositStatusPending,
	}

	if err := s.storage.CreateDeposit(deposit); err != nil {
		return "", fmt.Errorf("failed to save deposit: %w", err)
	}

	return address, nil
}

// GetBalance returns balance of hot wallet
func (s *WalletService) GetBalance(ctx context.Context, chain models.Chain) (*big.Int, error) {
	wallet, err := s.storage.GetHotWallet(chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get hot wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("hot wallet not found for chain %s", chain)
	}

	adapter, ok := s.adapters[chain]
	if !ok {
		return nil, fmt.Errorf("chain %s not supported", chain)
	}

	balance, err := adapter.GetBalance(ctx, wallet.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Update cached balance
	balanceStr := balance.String()
	if err := s.storage.UpdateHotWalletBalance(chain, balanceStr); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to update cached balance: %v\n", err)
	}

	return balance, nil
}

