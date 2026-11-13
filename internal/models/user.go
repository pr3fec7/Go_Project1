package models

// User — зарегистрированный пользователь.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
}
