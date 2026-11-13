package handlers

import (
	"context"
	"errors"
	"net/http"

	"expenses/internal/repository"
)

// contextKey — приватный тип для ключа в контексте.
// Делает невозможным случайное совпадение ключей из разных пакетов.
type contextKey string

const userIDKey contextKey = "userID"

type AuthMiddleware struct {
	sessions *repository.SessionRepository
}

func NewAuthMiddleware(sessions *repository.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{sessions: sessions}
}

// RequireAuth оборачивает http.Handler и проверяет авторизацию.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID, err := m.sessions.GetUserID(cookie.Value)
		if errors.Is(err, repository.ErrNotFound) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if err != nil {
			http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID достаёт userID из контекста запроса.
// Вызывается ТОЛЬКО в защищённых обработчиках.
func GetUserID(r *http.Request) (int64, bool) {
	id, ok := r.Context().Value(userIDKey).(int64)
	return id, ok
}
