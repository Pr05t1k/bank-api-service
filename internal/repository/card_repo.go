package repository

import (
	"database/sql"

	"bank-api/internal/models"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Create(card *models.Card) error {
	query := `
        INSERT INTO cards (account_id, number, number_hmac, expiry_month, expiry_year, cvv_hash, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at`

	err := r.db.QueryRow(query, card.AccountID, card.Number, card.NumberHMAC,
		card.ExpiryMonth, card.ExpiryYear, card.CVVHash, card.CreatedAt).
		Scan(&card.ID, &card.CreatedAt)

	return err
}

func (r *CardRepository) FindByAccountID(accountID int) ([]models.Card, error) {
	rows, err := r.db.Query(
		`SELECT id, account_id, number, number_hmac, expiry_month, expiry_year, cvv_hash, created_at
         FROM cards WHERE account_id = $1`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		err := rows.Scan(&card.ID, &card.AccountID, &card.Number, &card.NumberHMAC,
			&card.ExpiryMonth, &card.ExpiryYear, &card.CVVHash, &card.CreatedAt)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (r *CardRepository) FindByID(id int) (*models.Card, error) {
	card := &models.Card{}
	query := `SELECT id, account_id, number, number_hmac, expiry_month, expiry_year, cvv_hash, created_at
              FROM cards WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(&card.ID, &card.AccountID, &card.Number,
		&card.NumberHMAC, &card.ExpiryMonth, &card.ExpiryYear, &card.CVVHash, &card.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, models.ErrCardNotFound
	}
	return card, err
}
