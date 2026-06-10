package service

import (
	"time"

	"bank-api/internal/models"
	"bank-api/internal/repository"
	"bank-api/internal/utils"

	"github.com/sirupsen/logrus"
)

type CreditService struct {
	creditRepo         *repository.CreditRepository
	accountRepo        *repository.AccountRepository
	transactionRepo    *repository.TransactionRepository
	getCentralBankRate func() (float64, error)
	logger             *logrus.Logger
}

func NewCreditService(
	creditRepo *repository.CreditRepository,
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	getCentralBankRate func() (float64, error),
	logger *logrus.Logger,
) *CreditService {
	return &CreditService{
		creditRepo:         creditRepo,
		accountRepo:        accountRepo,
		transactionRepo:    transactionRepo,
		getCentralBankRate: getCentralBankRate,
		logger:             logger,
	}
}

// ApplyForCredit оформляет кредит
func (s *CreditService) ApplyForCredit(accountID int, amount float64, termMonths int, userID int) (*models.CreditResponse, error) {
	// Проверяем счет
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}

	if account.UserID != userID {
		return nil, models.ErrAccountNotFound
	}

	// Получаем ключевую ставку ЦБ РФ и добавляем маржу банка (+5%)
	centralRate, err := s.getCentralBankRate()
	if err != nil {
		s.logger.Warn("Failed to get central bank rate, using default: ", err)
		centralRate = 7.5 // default rate
	}

	interestRate := centralRate + 5 // +5% маржа банка

	// Рассчитываем аннуитетный платеж
	monthlyPayment := utils.CalculateAnnuityPayment(amount, interestRate, termMonths)
	totalPayment := monthlyPayment * float64(termMonths)

	// Создаем кредит
	credit := &models.Credit{
		AccountID:       accountID,
		Amount:          amount,
		InterestRate:    interestRate,
		TermMonths:      termMonths,
		MonthlyPayment:  monthlyPayment,
		RemainingAmount: amount,
		Status:          "active",
	}

	if err := s.creditRepo.Create(credit); err != nil {
		s.logger.Error("Failed to create credit: ", err)
		return nil, err
	}

	// Зачисляем деньги на счет
	newBalance := account.Balance + amount
	if err := s.accountRepo.UpdateBalance(accountID, newBalance); err != nil {
		return nil, err
	}

	// Создаем график платежей
	schedule := utils.CalculatePaymentSchedule(amount, interestRate, termMonths, monthlyPayment)
	dueDate := time.Now().AddDate(0, 1, 0) // первый платеж через месяц

	for i, payment := range schedule {
		ps := &models.PaymentSchedule{
			CreditID:      credit.ID,
			PaymentNumber: payment.Number,
			DueDate:       dueDate.AddDate(0, i, 0),
			Amount:        payment.Principal + payment.Interest,
			Principal:     payment.Principal,
			Interest:      payment.Interest,
			Status:        "pending",
		}
		if err := s.creditRepo.CreatePaymentSchedule(ps); err != nil {
			s.logger.Error("Failed to create payment schedule: ", err)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"credit_id":       credit.ID,
		"amount":          amount,
		"monthly_payment": monthlyPayment,
	}).Info("Credit approved")

	return &models.CreditResponse{
		ID:              credit.ID,
		Amount:          amount,
		InterestRate:    interestRate,
		TermMonths:      termMonths,
		MonthlyPayment:  monthlyPayment,
		TotalPayment:    totalPayment,
		RemainingAmount: amount,
		Status:          "active",
		CreatedAt:       credit.CreatedAt,
	}, nil
}

// ProcessOverduePayments обрабатывает просроченные платежи (для шедулера)
func (s *CreditService) ProcessOverduePayments() error {
	s.logger.Info("Processing overdue payments")

	overduePayments, err := s.creditRepo.GetOverduePayments()
	if err != nil {
		return err
	}

	for _, payment := range overduePayments {
		credit, err := s.creditRepo.FindByID(payment.CreditID)
		if err != nil {
			continue
		}

		account, err := s.accountRepo.FindByID(credit.AccountID)
		if err != nil {
			continue
		}

		penaltyAmount := utils.CalculateOverduePenalty(payment.Amount)

		if account.Balance >= penaltyAmount {
			// Списываем с пеней
			newBalance := account.Balance - penaltyAmount
			s.accountRepo.UpdateBalance(credit.AccountID, newBalance)

			// Обновляем статус платежа
			now := time.Now()
			s.creditRepo.UpdatePaymentStatus(payment.ID, "overdue", &now)

			// Обновляем остаток по кредиту
			newRemaining := credit.RemainingAmount - payment.Principal
			s.creditRepo.UpdateRemainingAmount(credit.ID, newRemaining)

			s.logger.WithFields(logrus.Fields{
				"credit_id":  credit.ID,
				"payment_id": payment.ID,
				"penalty":    penaltyAmount - payment.Amount,
			}).Info("Overdue payment processed with penalty")
		} else {
			s.logger.WithFields(logrus.Fields{
				"credit_id":  credit.ID,
				"payment_id": payment.ID,
				"needed":     penaltyAmount,
				"available":  account.Balance,
			}).Warn("Insufficient funds for overdue payment")
		}
	}

	return nil
}

// GetPaymentSchedule возвращает график платежей по кредиту
func (s *CreditService) GetPaymentSchedule(creditID, userID int) ([]models.PaymentSchedule, error) {
	credit, err := s.creditRepo.FindByID(creditID)
	if err != nil {
		return nil, err
	}

	account, err := s.accountRepo.FindByID(credit.AccountID)
	if err != nil {
		return nil, err
	}

	if account.UserID != userID {
		return nil, models.ErrCreditNotFound
	}

	return s.creditRepo.GetPaymentSchedule(creditID)
}
