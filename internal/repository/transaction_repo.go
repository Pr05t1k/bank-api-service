package repository

import (
	"database/sql"
	"time"

	"bank-api/internal/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (from_account_id, to_account_id, amount, type, status, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	tx.CreatedAt = time.Now()
	tx.Status = "pending"

	err := r.db.QueryRow(query, tx.FromAccountID, tx.ToAccountID, tx.Amount, tx.Type, tx.Status, tx.Description, tx.CreatedAt).
		Scan(&tx.ID)

	return err
}

func (r *TransactionRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE transactions SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *TransactionRepository) GetByAccountID(accountID int) ([]models.Transaction, error) {
	rows, err := r.db.Query(
		`SELECT id, from_account_id, to_account_id, amount, type, status, description, created_at
		 FROM transactions 
		 WHERE from_account_id = $1 OR to_account_id = $1
		 ORDER BY created_at DESC`, accountID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(&tx.ID, &tx.FromAccountID, &tx.ToAccountID, &tx.Amount, &tx.Type, &tx.Status, &tx.Description, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}
