package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/service"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AccountHandler struct {
	accountService *service.AccountService
	logger         *logrus.Logger
}

func NewAccountHandler(accountService *service.AccountService, logger *logrus.Logger) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		logger:         logger,
	}
}

// CreateAccount создает новый счет
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста (JWT)
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.accountService.CreateAccount(userID)
	if err != nil {
		h.logger.Error("Failed to create account: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

// GetAccounts возвращает все счета пользователя
func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accounts, err := h.accountService.GetUserAccounts(userID)
	if err != nil {
		h.logger.Error("Failed to get accounts: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
}

// GetAccountByID возвращает счет по ID
func (h *AccountHandler) GetAccountByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	account, err := h.accountService.GetAccountByID(accountID, userID)
	if err != nil {
		h.logger.Error("Failed to get account: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(account)
}
