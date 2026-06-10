package handlers

import (
	"encoding/json"
	"net/http"

	"bank-api/internal/models"
	"bank-api/internal/service"

	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	userService *service.UserService
	logger      *logrus.Logger
}

func NewUserHandler(userService *service.UserService, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// Register обрабатывает регистрацию пользователя
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email and password are required", http.StatusBadRequest)
		return
	}

	// Регистрация
	user, err := h.userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Registration failed: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login обрабатывает вход пользователя
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Логин
	token, user, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		h.logger.Error("Login failed: ", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.LoginResponse{
		Token: token,
		User:  user,
	})
}
