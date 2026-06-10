package repository

import (
	"database/sql"
	"time"

	"bank-api/internal/models"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

// Credit methods
func (r *CreditRepository) Create(credit *models.Credit) error {
	query := `
        INSERT INTO credits (account_id, amount, interest_rate, term_months, monthly_payment, remaining_amount, status, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id`

	credit.CreatedAt = time.Now()
	credit.Status = "active"

	err := r.db.QueryRow(query, credit.AccountID, credit.Amount, credit.InterestRate,
		credit.TermMonths, credit.MonthlyPayment, credit.RemainingAmount, credit.Status, credit.CreatedAt).
		Scan(&credit.ID)

	return err
}

func (r *CreditRepository) FindByID(id int) (*models.Credit, error) {
	credit := &models.Credit{}
	query := `SELECT id, account_id, amount, interest_rate, term_months, monthly_payment, remaining_amount, status, created_at
              FROM credits WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(&credit.ID, &credit.AccountID, &credit.Amount,
		&credit.InterestRate, &credit.TermMonths, &credit.MonthlyPayment,
		&credit.RemainingAmount, &credit.Status, &credit.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, models.ErrCreditNotFound
	}
	return credit, err
}

func (r *CreditRepository) FindByAccountID(accountID int) ([]models.Credit, error) {
	rows, err := r.db.Query(
		`SELECT id, account_id, amount, interest_rate, term_months, monthly_payment, remaining_amount, status, created_at
         FROM credits WHERE account_id = $1 ORDER BY created_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []models.Credit
	for rows.Next() {
		var credit models.Credit
		err := rows.Scan(&credit.ID, &credit.AccountID, &credit.Amount,
			&credit.InterestRate, &credit.TermMonths, &credit.MonthlyPayment,
			&credit.RemainingAmount, &credit.Status, &credit.CreatedAt)
		if err != nil {
			return nil, err
		}
		credits = append(credits, credit)
	}
	return credits, nil
}

func (r *CreditRepository) UpdateRemainingAmount(id int, remainingAmount float64) error {
	query := `UPDATE credits SET remaining_amount = $1 WHERE id = $2`
	_, err := r.db.Exec(query, remainingAmount, id)
	return err
}

func (r *CreditRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE credits SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

// PaymentSchedule methods
func (r *CreditRepository) CreatePaymentSchedule(schedule *models.PaymentSchedule) error {
	query := `
        INSERT INTO payment_schedules (credit_id, payment_number, due_date, amount, principal, interest, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`

	err := r.db.QueryRow(query, schedule.CreditID, schedule.PaymentNumber,
		schedule.DueDate, schedule.Amount, schedule.Principal, schedule.Interest, schedule.Status).
		Scan(&schedule.ID)

	return err
}

func (r *CreditRepository) GetPaymentSchedule(creditID int) ([]models.PaymentSchedule, error) {
	rows, err := r.db.Query(
		`SELECT id, credit_id, payment_number, due_date, amount, principal, interest, status, paid_at
         FROM payment_schedules WHERE credit_id = $1 ORDER BY payment_number`, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.PaymentSchedule
	for rows.Next() {
		var ps models.PaymentSchedule
		err := rows.Scan(&ps.ID, &ps.CreditID, &ps.PaymentNumber, &ps.DueDate,
			&ps.Amount, &ps.Principal, &ps.Interest, &ps.Status, &ps.PaidAt)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, ps)
	}
	return schedules, nil
}

func (r *CreditRepository) UpdatePaymentStatus(id int, status string, paidAt *time.Time) error {
	query := `UPDATE payment_schedules SET status = $1, paid_at = $2 WHERE id = $3`
	_, err := r.db.Exec(query, status, paidAt, id)
	return err
}

func (r *CreditRepository) GetOverduePayments() ([]models.PaymentSchedule, error) {
	rows, err := r.db.Query(
		`SELECT ps.id, ps.credit_id, ps.payment_number, ps.due_date, ps.amount, ps.principal, ps.interest, ps.status, ps.paid_at,
                c.account_id, c.interest_rate
         FROM payment_schedules ps
         JOIN credits c ON ps.credit_id = c.id
         WHERE ps.status = 'pending' AND ps.due_date < NOW()`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.PaymentSchedule
	for rows.Next() {
		var ps models.PaymentSchedule
		var accountID int
		var interestRate float64
		err := rows.Scan(&ps.ID, &ps.CreditID, &ps.PaymentNumber, &ps.DueDate,
			&ps.Amount, &ps.Principal, &ps.Interest, &ps.Status, &ps.PaidAt,
			&accountID, &interestRate)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, ps)
	}
	return schedules, nil
}
