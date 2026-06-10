package repository

import (
	"database/sql"

	"bank-api/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRow(query, user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return models.ErrDuplicateUser
	}
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, created_at 
			  FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.PasswordHash, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	return user, err
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password_hash, created_at 
			  FROM users WHERE username = $1`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.PasswordHash, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, models.ErrUserNotFound
	}
	return user, err
}
