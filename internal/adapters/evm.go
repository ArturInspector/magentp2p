package adapters

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EVMAdapter implements BlockchainAdapter for EVM-compatible chains
type EVMAdapter struct {
	client  *ethclient.Client
	chainID *big.Int
}

func NewEVMAdapter(rpcURL string, chainID int64) (*EVMAdapter, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	return &EVMAdapter{
		client:  client,
		chainID: big.NewInt(chainID),
	}, nil
}

func (e *EVMAdapter) GenerateAddress(ctx context.Context) (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("failed to cast public key")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address.Hex(), nil
}

func (e *EVMAdapter) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	balance, err := e.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

func (e *EVMAdapter) SendTransaction(ctx context.Context, fromAddress, toAddress string, amount *big.Int, privateKeyHex string) (string, error) {
	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	fromAddr := common.HexToAddress(fromAddress)
	toAddr := common.HexToAddress(toAddress)

	// Get nonce
	nonce, err := e.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, toAddr, amount, 21000, gasPrice, nil)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(e.chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %w", err)
	}

	// Send transaction
	err = e.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send tx: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

func (e *EVMAdapter) GetTransactionStatus(ctx context.Context, txHash string) (*TransactionStatus, error) {
	hash := common.HexToHash(txHash)
	_, isPending, err := e.client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx: %w", err)
	}

	if isPending {
		return &TransactionStatus{
			Status:       "pending",
			BlockNumber:  0,
			Confirmations: 0,
			Success:      false,
		}, nil
	}

	receipt, err := e.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}

	// Get latest block
	latestBlock, err := e.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	confirmations := int(latestBlock - receipt.BlockNumber.Uint64())

	return &TransactionStatus{
		Status:       "confirmed",
		BlockNumber:  receipt.BlockNumber.Int64(),
		Confirmations: confirmations,
		Success:      receipt.Status == 1,
	}, nil
}

func (e *EVMAdapter) GetLatestBlock(ctx context.Context) (int64, error) {
	blockNum, err := e.client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest block: %w", err)
	}
	return int64(blockNum), nil
}

func (e *EVMAdapter) GetBlockTransactions(ctx context.Context, blockNumber int64) ([]*Transaction, error) {
	block, err := e.client.BlockByNumber(ctx, big.NewInt(blockNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	var transactions []*Transaction
	for _, tx := range block.Transactions() {
		// Only process regular transactions (not contract creation)
		if tx.To() == nil {
			continue
		}

		transactions = append(transactions, &Transaction{
			Hash:     tx.Hash().Hex(),
			From:     "", // TODO: get from address
			To:       tx.To().Hex(),
			Amount:   tx.Value(),
			BlockNum: blockNumber,
		})
	}

	return transactions, nil
}

func (e *EVMAdapter) GetGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := e.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	return gasPrice, nil
}

func (e *EVMAdapter) Close() {
	if e.client != nil {
		e.client.Close()
	}
}

