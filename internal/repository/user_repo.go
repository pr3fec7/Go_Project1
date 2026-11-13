package repository

import (
	"database/sql"
	"errors"
	"strings"

	"expenses/internal/models"
)

var ErrEmailTaken = errors.New("email already taken")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create регистрирует пользователя. Возвращает ID или ErrEmailTaken.
func (r *UserRepository) Create(email, passwordHash string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		email, passwordHash,
	)
	if err != nil {
		// SQLite не даёт нормального кода ошибки — проверяем по тексту
		if isUniqueViolation(err) {
			return 0, ErrEmailTaken
		}
		return 0, err
	}
	return res.LastInsertId()
}

// GetByEmail находит пользователя для входа.
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(
		`SELECT id, email, password_hash FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
