package service

import (
	"strconv"

	"bank-api/internal/middleware"
	"bank-api/internal/models"
	"bank-api/internal/repository"
	"bank-api/internal/utils"

	"github.com/sirupsen/logrus"
)

type UserService struct {
	userRepo *repository.UserRepository
	logger   *logrus.Logger
}

func NewUserService(userRepo *repository.UserRepository, logger *logrus.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *UserService) Register(username, email, password string) (*models.User, error) {
	s.logger.WithFields(logrus.Fields{
		"username":        username,
		"email":           email,
		"password_length": len(password),
	}).Info("Attempting user registration")

	// Хешируем пароль
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		s.logger.WithError(err).Error("Failed to hash password")
		return nil, err
	}

	s.logger.WithField("hash", hashedPassword[:20]).Info("Password hashed successfully")

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
	}

	// Валидация
	if err := user.Validate(); err != nil {
		s.logger.WithError(err).Warn("Validation failed")
		return nil, err
	}

	// Сохраняем в БД
	if err := s.userRepo.Create(user); err != nil {
		s.logger.WithError(err).Error("Failed to create user in database")
		return nil, err
	}

	s.logger.WithField("user_id", user.ID).Info("User registered successfully")
	return user, nil
}

func (s *UserService) Login(email, password string) (string, *models.User, error) {
	s.logger.WithField("email", email).Info("Login attempt")

	// Находим пользователя
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		s.logger.WithError(err).Warn("User not found")
		return "", nil, models.ErrInvalidPassword
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     user.ID,
		"username":    user.Username,
		"hash_length": len(user.PasswordHash),
	}).Info("User found, checking password")

	// Проверяем пароль
	passwordValid := utils.CheckPasswordHash(password, user.PasswordHash)
	s.logger.WithField("password_valid", passwordValid).Info("Password check completed")

	if !passwordValid {
		s.logger.Warn("Invalid password for user: ", email)
		return "", nil, models.ErrInvalidPassword
	}

	// Генерируем JWT токен
	token, err := middleware.GenerateJWT(strconv.Itoa(user.ID))
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate JWT")
		return "", nil, err
	}

	s.logger.WithField("user_id", user.ID).Info("User logged in successfully")
	return token, user, nil
}
