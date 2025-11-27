package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/dechat/exchange-service/internal/adapters"
	"github.com/dechat/exchange-service/internal/models"
	"github.com/dechat/exchange-service/internal/storage"
)

type WithdrawalService struct {
	storage  storage.Storage
	adapters map[models.Chain]adapters.BlockchainAdapter
	encKey   []byte // TODO: load from secure config
}

func NewWithdrawalService(storage storage.Storage, adapters map[models.Chain]adapters.BlockchainAdapter) *WithdrawalService {
	// TODO: load encryption key from secure config
	encKey := make([]byte, 32)
	rand.Read(encKey)

	return &WithdrawalService{
		storage:  storage,
		adapters: adapters,
		encKey:   encKey,
	}
}

// ProcessPendingWithdrawals processes pending withdrawals
func (s *WithdrawalService) ProcessPendingWithdrawals(ctx context.Context, chain models.Chain) error {
	withdrawals, err := s.storage.GetPendingWithdrawals(chain, 10)
	if err != nil {
		return fmt.Errorf("failed to get pending withdrawals: %w", err)
	}

	adapter, ok := s.adapters[chain]
	if !ok {
		return fmt.Errorf("chain %s not supported", chain)
	}

	for _, withdrawal := range withdrawals {
		if err := s.processWithdrawal(ctx, adapter, withdrawal); err != nil {
			fmt.Printf("Error processing withdrawal %d: %v\n", withdrawal.ID, err)
			// Continue with next withdrawal
			continue
		}
	}

	return nil
}

func (s *WithdrawalService) processWithdrawal(ctx context.Context, adapter adapters.BlockchainAdapter, withdrawal *models.Withdrawal) error {
	// Get hot wallet
	wallet, err := s.storage.GetHotWallet(withdrawal.Chain)
	if err != nil {
		return fmt.Errorf("failed to get hot wallet: %w", err)
	}
	if wallet == nil {
		return fmt.Errorf("hot wallet not found for chain %s", withdrawal.Chain)
	}

	// Check balance
	balance, err := adapter.GetBalance(ctx, wallet.Address)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	amount, ok := new(big.Int).SetString(withdrawal.Amount, 10)
	if !ok {
		return fmt.Errorf("invalid amount: %s", withdrawal.Amount)
	}

	// Get gas price
	gasPrice, err := adapter.GetGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate fee (21000 gas for simple transfer)
	gasLimit := big.NewInt(21000)
	fee := new(big.Int).Mul(gasPrice, gasLimit)

	// Check if balance is sufficient
	totalNeeded := new(big.Int).Add(amount, fee)
	if balance.Cmp(totalNeeded) < 0 {
		return fmt.Errorf("insufficient balance: have %s, need %s", balance.String(), totalNeeded.String())
	}

	// Decrypt private key
	privateKey, err := s.decryptKey(wallet.EncryptedKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt key: %w", err)
	}

	// Send transaction
	txHash, err := adapter.SendTransaction(ctx, wallet.Address, withdrawal.ToAddress, amount, privateKey)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	// Update withdrawal
	withdrawal.TxHash = txHash
	withdrawal.Status = models.WithdrawalStatusSent
	withdrawal.Fee = fee.String()
	now := time.Now()
	withdrawal.SentAt = &now

	if err := s.storage.UpdateWithdrawal(withdrawal); err != nil {
		return fmt.Errorf("failed to update withdrawal: %w", err)
	}

	fmt.Printf("Withdrawal sent: chain=%s, order_id=%s, tx_hash=%s\n", withdrawal.Chain, withdrawal.OrderID, txHash)
	return nil
}

// CreateWithdrawal creates new withdrawal request
func (s *WithdrawalService) CreateWithdrawal(ctx context.Context, chain models.Chain, orderID, toAddress, amount string) (*models.Withdrawal, error) {
	wallet, err := s.storage.GetHotWallet(chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get hot wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("hot wallet not found for chain %s", chain)
	}

	withdrawal := &models.Withdrawal{
		Chain:       chain,
		OrderID:     orderID,
		FromAddress: wallet.Address,
		ToAddress:   toAddress,
		Amount:      amount,
		Fee:         "0", // Will be calculated when sending
		Status:      models.WithdrawalStatusPending,
	}

	if err := s.storage.CreateWithdrawal(withdrawal); err != nil {
		return nil, fmt.Errorf("failed to create withdrawal: %w", err)
	}

	return withdrawal, nil
}

// Simple encryption/decryption (TODO: use proper key management)
func (s *WithdrawalService) encryptKey(key string) (string, error) {
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(key), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (s *WithdrawalService) decryptKey(encryptedHex string) (string, error) {
	ciphertext, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

