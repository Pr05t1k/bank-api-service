package service

import (
	"fmt"

	"bank-api/internal/models"
	"bank-api/internal/repository"

	"github.com/sirupsen/logrus"
)

type TransferService struct {
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	logger          *logrus.Logger
}

func NewTransferService(
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	logger *logrus.Logger,
) *TransferService {
	return &TransferService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

func (s *TransferService) Transfer(fromAccountID, toAccountID int, amount float64, userID int, description string) (*models.Transaction, error) {
	s.logger.WithFields(logrus.Fields{
		"from":   fromAccountID,
		"to":     toAccountID,
		"amount": amount,
	}).Info("Initiating transfer")

	fromAccount, err := s.accountRepo.FindByID(fromAccountID)
	if err != nil {
		return nil, err
	}

	if fromAccount.UserID != userID {
		return nil, fmt.Errorf("access denied to account %d", fromAccountID)
	}

	_, err = s.accountRepo.FindByID(toAccountID)
	if err != nil {
		return nil, fmt.Errorf("destination account not found")
	}

	transaction := &models.Transaction{
		FromAccountID: &fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Type:          "transfer",
		Description:   description,
	}

	if err := s.transactionRepo.Create(transaction); err != nil {
		s.logger.Error("Failed to create transaction: ", err)
		return nil, err
	}

	if err := s.accountRepo.Transfer(fromAccountID, toAccountID, amount); err != nil {
		transaction.Status = "failed"
		s.transactionRepo.UpdateStatus(transaction.ID, "failed")
		s.logger.Error("Transfer failed: ", err)
		return nil, err
	}

	transaction.Status = "completed"
	s.transactionRepo.UpdateStatus(transaction.ID, "completed")

	s.logger.WithFields(logrus.Fields{
		"transaction_id": transaction.ID,
		"from":           fromAccountID,
		"to":             toAccountID,
		"amount":         amount,
	}).Info("Transfer completed successfully")

	return transaction, nil
}

func (s *TransferService) GetTransactionHistory(accountID, userID int) ([]models.Transaction, error) {
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}

	if account.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	return s.transactionRepo.GetByAccountID(accountID)
}
