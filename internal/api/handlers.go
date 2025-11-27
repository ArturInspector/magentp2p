package api

import (
	"encoding/json"
	"net/http"

	"github.com/dechat/exchange-service/internal/models"
	"github.com/dechat/exchange-service/internal/services"
	"github.com/gorilla/mux"
)

type Handlers struct {
	walletService     *services.WalletService
	withdrawalService *services.WithdrawalService
}

func NewHandlers(walletService *services.WalletService, withdrawalService *services.WithdrawalService) *Handlers {
	return &Handlers{
		walletService:     walletService,
		withdrawalService: withdrawalService,
	}
}

// GenerateDepositAddressRequest request for deposit address generation
type GenerateDepositAddressRequest struct {
	Chain   string `json:"chain"`
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
}

// GenerateDepositAddressResponse response with deposit address
type GenerateDepositAddressResponse struct {
	Address string `json:"address"`
	Chain   string `json:"chain"`
	OrderID string `json:"order_id"`
}

func (h *Handlers) GenerateDepositAddress(w http.ResponseWriter, r *http.Request) {
	var req GenerateDepositAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	chain := models.Chain(req.Chain)
	address, err := h.walletService.GenerateDepositAddress(r.Context(), chain, req.UserID, req.OrderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := GenerateDepositAddressResponse{
		Address: address,
		Chain:   string(chain),
		OrderID: req.OrderID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetBalanceResponse response with balance
type GetBalanceResponse struct {
	Chain   string `json:"chain"`
	Balance string `json:"balance"`
}

func (h *Handlers) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chain := models.Chain(vars["chain"])

	balance, err := h.walletService.GetBalance(r.Context(), chain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := GetBalanceResponse{
		Chain:   string(chain),
		Balance: balance.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateWithdrawalRequest request for withdrawal
type CreateWithdrawalRequest struct {
	Chain    string `json:"chain"`
	OrderID  string `json:"order_id"`
	ToAddress string `json:"to_address"`
	Amount   string `json:"amount"`
}

func (h *Handlers) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	var req CreateWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	chain := models.Chain(req.Chain)
	withdrawal, err := h.withdrawalService.CreateWithdrawal(r.Context(), chain, req.OrderID, req.ToAddress, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(withdrawal)
}

// GetTransactionRequest request for transaction status
type GetTransactionRequest struct {
	Chain  string `json:"chain"`
	TxHash string `json:"tx_hash"`
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

