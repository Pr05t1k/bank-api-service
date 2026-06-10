package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"bank-api/internal/middleware"
	"bank-api/internal/service"
)

type CardHandler struct {
	cardService *service.CardService
	logger      *logrus.Logger
}

func NewCardHandler(cardService *service.CardService, logger *logrus.Logger) *CardHandler {
	return &CardHandler{
		cardService: cardService,
		logger:      logger,
	}
}

func (h *CardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["account_id"])
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	card, cvv, err := h.cardService.CreateCard(accountID, userID)
	if err != nil {
		h.logger.Error("Failed to create card: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"card": card,
		"cvv":  cvv, // В реальном проекте отправлять по email/SMS
	})
}

func (h *CardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.Atoi(vars["account_id"])
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cards, err := h.cardService.GetCards(accountID, userID)
	if err != nil {
		h.logger.Error("Failed to get cards: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}
