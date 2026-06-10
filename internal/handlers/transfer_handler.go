package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/service"

	"github.com/sirupsen/logrus"
)

type TransferHandler struct {
	transferService *service.TransferService
	logger          *logrus.Logger
}

func NewTransferHandler(transferService *service.TransferService, logger *logrus.Logger) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
		logger:          logger,
	}
}

func (h *TransferHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req models.TransferRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FromAccountID == 0 || req.ToAccountID == 0 {
		http.Error(w, "from_account_id and to_account_id are required", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.FromAccountID == req.ToAccountID {
		http.Error(w, "cannot transfer to the same account", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	transaction, err := h.transferService.Transfer(
		req.FromAccountID,
		req.ToAccountID,
		req.Amount,
		userID,
		req.Description,
	)

	if err != nil {
		h.logger.Error("Transfer failed: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransferHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	accountIDStr := r.URL.Query().Get("account_id")
	if accountIDStr == "" {
		http.Error(w, "account_id is required", http.StatusBadRequest)
		return
	}

	var accountID int
	_, err := fmt.Sscanf(accountIDStr, "%d", &accountID)
	if err != nil {
		http.Error(w, "invalid account_id", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	transactions, err := h.transferService.GetTransactionHistory(accountID, userID)
	if err != nil {
		h.logger.Error("Failed to get transaction history: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}
