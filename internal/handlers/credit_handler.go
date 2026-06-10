package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/service"
)

type CreditHandler struct {
	creditService *service.CreditService
	logger        *logrus.Logger
}

func NewCreditHandler(creditService *service.CreditService, logger *logrus.Logger) *CreditHandler {
	return &CreditHandler{
		creditService: creditService,
		logger:        logger,
	}
}

func (h *CreditHandler) ApplyForCredit(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCreditRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.TermMonths < 1 || req.TermMonths > 60 {
		http.Error(w, "Term must be between 1 and 60 months", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	credit, err := h.creditService.ApplyForCredit(req.AccountID, req.Amount, req.TermMonths, userID)
	if err != nil {
		h.logger.Error("Failed to apply for credit: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(credit)
}

func (h *CreditHandler) GetPaymentSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditID, err := strconv.Atoi(vars["credit_id"])
	if err != nil {
		http.Error(w, "Invalid credit ID", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	schedule, err := h.creditService.GetPaymentSchedule(creditID, userID)
	if err != nil {
		h.logger.Error("Failed to get payment schedule: ", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}
