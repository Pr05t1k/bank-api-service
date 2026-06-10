package service

import (
	"time"

	"bank-api/internal/models"
	"bank-api/internal/repository"
	"bank-api/internal/utils"

	"github.com/sirupsen/logrus"
)

type CardService struct {
	cardRepo    *repository.CardRepository
	accountRepo *repository.AccountRepository
	hmacSecret  []byte
	logger      *logrus.Logger
}

func NewCardService(
	cardRepo *repository.CardRepository,
	accountRepo *repository.AccountRepository,
	hmacSecret []byte,
	logger *logrus.Logger,
) *CardService {
	return &CardService{
		cardRepo:    cardRepo,
		accountRepo: accountRepo,
		hmacSecret:  hmacSecret,
		logger:      logger,
	}
}

func (s *CardService) CreateCard(accountID, userID int) (*models.CardResponse, string, error) {
	// Проверяем, что счет принадлежит пользователю
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, "", err
	}

	if account.UserID != userID {
		return nil, "", models.ErrAccountNotFound
	}

	// Генерируем номер карты
	cardNumber := utils.GenerateCardNumber("4") // Visa prefix
	if !utils.LuhnCheck(cardNumber) {
		s.logger.Warn("Generated card number failed Luhn check")
	}

	// Временное решение: для демо используем base64 вместо PGP
	// В реальном проекте используй полноценное PGP шифрование
	encryptedNumber := cardNumber // TODO: добавить PGP шифрование
	numberHMAC := utils.ComputeHMAC(cardNumber, s.hmacSecret)

	// Генерируем CVV и хешируем его
	cvv := utils.GenerateCVV()
	cvvHash, err := utils.HashCVV(cvv)
	if err != nil {
		return nil, "", err
	}

	// Создаем карту
	card := &models.Card{
		AccountID:   accountID,
		Number:      encryptedNumber,
		NumberHMAC:  numberHMAC,
		ExpiryMonth: int(time.Now().Month()),
		ExpiryYear:  time.Now().Year() + 3,
		CVVHash:     cvvHash,
		CreatedAt:   time.Now(),
	}

	if err := s.cardRepo.Create(card); err != nil {
		s.logger.Error("Failed to create card: ", err)
		return nil, "", err
	}

	s.logger.WithFields(logrus.Fields{
		"card_id":    card.ID,
		"account_id": accountID,
	}).Info("Card created successfully")

	response := &models.CardResponse{
		ID:          card.ID,
		Last4:       cardNumber[len(cardNumber)-4:],
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
	}

	return response, cvv, nil
}

func (s *CardService) GetCards(accountID, userID int) ([]models.CardResponse, error) {
	account, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}

	if account.UserID != userID {
		return nil, models.ErrAccountNotFound
	}

	cards, err := s.cardRepo.FindByAccountID(accountID)
	if err != nil {
		return nil, err
	}

	var responses []models.CardResponse
	for _, card := range cards {
		// Расшифровываем номер (временное решение)
		decryptedNumber := card.Number // TODO: добавить PGP расшифровку

		responses = append(responses, models.CardResponse{
			ID:          card.ID,
			Last4:       decryptedNumber[len(decryptedNumber)-4:],
			ExpiryMonth: card.ExpiryMonth,
			ExpiryYear:  card.ExpiryYear,
		})
	}

	return responses, nil
}
