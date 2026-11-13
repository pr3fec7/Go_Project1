package models

import "time"

// Session — сессия авторизованного пользователя.
type Session struct {
	ID        string
	UserID    int64
	ExpiresAt time.Time
}
