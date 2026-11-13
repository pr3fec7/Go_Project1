package handlers

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"expenses/internal/repository"
)

type AuthHandler struct {
	users    *repository.UserRepository
	sessions *repository.SessionRepository

	loginTmpl    *template.Template
	registerTmpl *template.Template
}

func NewAuthHandler(
	users *repository.UserRepository,
	sessions *repository.SessionRepository,
) *AuthHandler {
	layout := filepath.Join("web", "templates", "layout.html")
	login := filepath.Join("web", "templates", "login.html")
	register := filepath.Join("web", "templates", "register.html")

	return &AuthHandler{
		users:        users,
		sessions:     sessions,
		loginTmpl:    template.Must(template.ParseFiles(layout, login)),
		registerTmpl: template.Must(template.ParseFiles(layout, register)),
	}
}

// RegisterPage — GET показывает форму, POST создаёт пользователя.
func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.renderRegister(w, "", "")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || len(password) < 6 {
		h.renderRegister(w, email, "Укажите email и пароль не короче 6 символов.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	_, err = h.users.Create(email, string(hash))
	if errors.Is(err, repository.ErrEmailTaken) {
		h.renderRegister(w, email, "Этот email уже зарегистрирован.")
		return
	}
	if err != nil {
		log.Printf("users.Create: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) renderRegister(w http.ResponseWriter, email, errMsg string) {
	data := map[string]any{
		"Title": "Регистрация",
		"Email": email,
		"Error": errMsg,
	}
	if err := h.registerTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("register template: %v", err)
	}
}

// LoginPage — GET показывает форму, POST проверяет пароль и создаёт сессию.
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.renderLogin(w, "", "")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	user, err := h.users.GetByEmail(email)
	if errors.Is(err, repository.ErrNotFound) {
		h.renderLogin(w, email, "Неверный email или пароль.")
		return
	}
	if err != nil {
		log.Printf("users.GetByEmail: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		h.renderLogin(w, email, "Неверный email или пароль.")
		return
	}

	session, err := h.sessions.Create(user.ID)
	if err != nil {
		log.Printf("sessions.Create: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 3600,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) renderLogin(w http.ResponseWriter, email, errMsg string) {
	data := map[string]any{
		"Title": "Вход",
		"Email": email,
		"Error": errMsg,
	}
	if err := h.loginTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("login template: %v", err)
	}
}

// Logout удаляет сессию и cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		_ = h.sessions.Delete(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
