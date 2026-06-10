package service

import (
	"errors"

	"bank-api/internal/models"
	"bank-api/internal/repository"

	"github.com/sirupsen/logrus"
)

type AccountService struct {
	accountRepo *repository.AccountRepository
	logger      *logrus.Logger
}

func NewAccountService(accountRepo *repository.AccountRepository, logger *logrus.Logger) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
		logger:      logger,
	}
}

func (s *AccountService) CreateAccount(userID int) (*models.Account, error) {
	account := &models.Account{
		UserID:  userID,
		Balance: 0,
	}

	if err := s.accountRepo.Create(account); err != nil {
		s.logger.Error("Failed to create account: ", err)
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": account.ID,
	}).Info("Account created successfully")

	return account, nil
}

func (s *AccountService) GetUserAccounts(userID int) ([]models.Account, error) {
	accounts, err := s.accountRepo.FindByUserID(userID)
	if err != nil {
		s.logger.Error("Failed to get user accounts: ", err)
		return nil, err
	}

	return accounts, nil
}

func (s *AccountService) GetAccountByID(accountID, userID int) (*models.Account, error) {
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}

	// Проверяем права доступа
	if account.UserID != userID {
		return nil, errors.New("access denied")
	}

	return account, nil
}
