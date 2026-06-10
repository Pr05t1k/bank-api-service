package models

import (
	"errors"
	"time"
)

var (
	ErrCardNotFound = errors.New("card not found")
)

type Card struct {
	ID          int       `json:"id" db:"id"`
	AccountID   int       `json:"account_id" db:"account_id"`
	Number      string    `json:"number" db:"number"` // зашифрован PGP
	NumberHMAC  string    `json:"-" db:"number_hmac"` // для проверки целостности
	ExpiryMonth int       `json:"expiry_month" db:"expiry_month"`
	ExpiryYear  int       `json:"expiry_year" db:"expiry_year"`
	CVVHash     string    `json:"-" db:"cvv_hash"` // bcrypt хеш
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type CreateCardRequest struct {
	AccountID int `json:"account_id"`
}

type CardResponse struct {
	ID          int    `json:"id"`
	Last4       string `json:"last4"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}
