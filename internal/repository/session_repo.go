package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"expenses/internal/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create создаёт новую сессию на 7 дней.
func (r *SessionRepository) Create(userID int64) (*models.Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = r.db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		sessionID, userID, expiresAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}

	return &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

// GetUserID возвращает ID пользователя по session_id, если сессия валидна.
func (r *SessionRepository) GetUserID(sessionID string) (int64, error) {
	var userID int64
	var expiresStr string
	err := r.db.QueryRow(
		`SELECT user_id, expires_at FROM sessions WHERE id = ?`, sessionID,
	).Scan(&userID, &expiresStr)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		return 0, err
	}
	if time.Now().After(expiresAt) {
		_ = r.Delete(sessionID) // просрочка — удаляем
		return 0, ErrNotFound
	}
	return userID, nil
}

// Delete удаляет сессию (выход из системы).
func (r *SessionRepository) Delete(sessionID string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

// generateSessionID создаёт случайный 32-байтный ID.
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
