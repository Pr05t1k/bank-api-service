package models

import (
	"errors"
	"time"
)

var (
	ErrCreditNotFound  = errors.New("credit not found")
	ErrPaymentNotFound = errors.New("payment not found")
)

type Credit struct {
	ID              int       `json:"id" db:"id"`
	AccountID       int       `json:"account_id" db:"account_id"`
	Amount          float64   `json:"amount" db:"amount"`
	InterestRate    float64   `json:"interest_rate" db:"interest_rate"` // годовая ставка
	TermMonths      int       `json:"term_months" db:"term_months"`
	MonthlyPayment  float64   `json:"monthly_payment" db:"monthly_payment"`
	RemainingAmount float64   `json:"remaining_amount" db:"remaining_amount"`
	Status          string    `json:"status" db:"status"` // active, paid, overdue
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type PaymentSchedule struct {
	ID            int        `json:"id" db:"id"`
	CreditID      int        `json:"credit_id" db:"credit_id"`
	PaymentNumber int        `json:"payment_number" db:"payment_number"`
	DueDate       time.Time  `json:"due_date" db:"due_date"`
	Amount        float64    `json:"amount" db:"amount"`
	Principal     float64    `json:"principal" db:"principal"`
	Interest      float64    `json:"interest" db:"interest"`
	Status        string     `json:"status" db:"status"` // pending, paid, overdue
	PaidAt        *time.Time `json:"paid_at,omitempty" db:"paid_at"`
}

type CreateCreditRequest struct {
	AccountID  int     `json:"account_id"`
	Amount     float64 `json:"amount"`
	TermMonths int     `json:"term_months"`
}

type CreditResponse struct {
	ID              int       `json:"id"`
	Amount          float64   `json:"amount"`
	InterestRate    float64   `json:"interest_rate"`
	TermMonths      int       `json:"term_months"`
	MonthlyPayment  float64   `json:"monthly_payment"`
	TotalPayment    float64   `json:"total_payment"`
	RemainingAmount float64   `json:"remaining_amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
