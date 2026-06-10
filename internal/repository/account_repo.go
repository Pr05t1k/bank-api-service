package repository

import (
	"database/sql"
	"fmt"
	"time"

	"bank-api/internal/models"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(account *models.Account) error {
	account.Number = generateAccountNumber()
	account.CreatedAt = time.Now()

	query := `
		INSERT INTO accounts (user_id, number, balance, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := r.db.QueryRow(query, account.UserID, account.Number, account.Balance, account.CreatedAt).
		Scan(&account.ID)

	return err
}

func (r *AccountRepository) FindByUserID(userID int) ([]models.Account, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, number, balance, created_at 
		 FROM accounts WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		err := rows.Scan(&acc.ID, &acc.UserID, &acc.Number, &acc.Balance, &acc.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (r *AccountRepository) FindByID(id int) (*models.Account, error) {
	account := &models.Account{}
	query := `SELECT id, user_id, number, balance, created_at 
			  FROM accounts WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&account.ID, &account.UserID, &account.Number,
		&account.Balance, &account.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, models.ErrAccountNotFound
	}
	return account, err
}

func (r *AccountRepository) UpdateBalance(accountID int, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	_, err := r.db.Exec(query, newBalance, accountID)
	return err
}

func (r *AccountRepository) GetBalance(accountID int) (float64, error) {
	var balance float64
	query := `SELECT balance FROM accounts WHERE id = $1`
	err := r.db.QueryRow(query, accountID).Scan(&balance)
	return balance, err
}

func (r *AccountRepository) Transfer(fromID, toID int, amount float64) error {
	// Начинаем транзакцию
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверяем баланс отправителя
	var fromBalance float64
	err = tx.QueryRow(`SELECT balance FROM accounts WHERE id = $1 FOR UPDATE`, fromID).Scan(&fromBalance)
	if err != nil {
		return err
	}

	if fromBalance < amount {
		return models.ErrInsufficientFunds
	}

	// Списываем сумму со счета отправителя
	_, err = tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromID)
	if err != nil {
		return err
	}

	// Зачисляем сумму на счет получателя
	_, err = tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toID)
	if err != nil {
		return err
	}

	// Фиксируем транзакцию
	return tx.Commit()
}

func generateAccountNumber() string {
	return fmt.Sprintf("40817%d", time.Now().UnixNano()%10000000000)
}
