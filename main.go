package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"expenses/internal/handlers"
	"expenses/internal/repository"
)

func main() {
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatalf("не удалось открыть БД: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("БД недоступна: %v", err)
	}

	if err := applyMigrations(db); err != nil {
		log.Fatalf("не удалось применить миграции: %v", err)
	}

	// Репозитории
	expenseRepo := repository.NewExpenseRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Обработчики
	expenseHandler := handlers.NewExpenseHandler(expenseRepo, categoryRepo)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo)

	// Middleware
	authMiddleware := handlers.NewAuthMiddleware(sessionRepo)

	mux := http.NewServeMux()

	// Публичные маршруты (не требуют авторизации)
	mux.HandleFunc("/register", authHandler.RegisterPage)
	mux.HandleFunc("/login", authHandler.LoginPage)
	mux.HandleFunc("/logout", authHandler.Logout)

	// Защищённые маршруты (требуют авторизации)
	protected := http.NewServeMux()
	protected.HandleFunc("/", expenseHandler.ListPage)
	protected.HandleFunc("/expenses/new", expenseHandler.NewExpensePage)
	protected.HandleFunc("/expenses", expenseHandler.CreateExpense)
	protected.HandleFunc("/expenses/edit", expenseHandler.EditPage)
	protected.HandleFunc("/expenses/update", expenseHandler.UpdateExpense)
	protected.HandleFunc("/expenses/delete", expenseHandler.DeleteExpense)

	mux.Handle("/", authMiddleware.RequireAuth(protected))

	// Статика
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := ":8080"
	log.Printf("Сервер запущен: http://localhost%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func applyMigrations(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}

	if version < 1 {
		if err := execFile(db, "migrations/001_init.sql"); err != nil {
			return fmt.Errorf("миграция 001: %w", err)
		}
		if _, err := db.Exec("PRAGMA user_version = 1"); err != nil {
			return err
		}
		version = 1
	}

	if version < 2 {
		if err := execFile(db, "migrations/002_categories.sql"); err != nil {
			return fmt.Errorf("миграция 002: %w", err)
		}
		if _, err := db.Exec("PRAGMA user_version = 2"); err != nil {
			return err
		}
		version = 2
	}

	if version < 3 {
		if err := execFile(db, "migrations/003_users_auth.sql"); err != nil {
			return fmt.Errorf("миграция 003: %w", err)
		}
		if _, err := db.Exec("PRAGMA user_version = 3"); err != nil {
			return err
		}
	}

	return nil
}

func execFile(db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(data))
	return err
}
