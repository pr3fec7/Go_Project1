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

	expenseRepo := repository.NewExpenseRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	h := handlers.NewExpenseHandler(expenseRepo, categoryRepo)

	mux := http.NewServeMux()

	mux.HandleFunc("/", h.ListPage)
	mux.HandleFunc("/expenses/new", h.NewExpensePage)
	mux.HandleFunc("/expenses", h.CreateExpense)
	mux.HandleFunc("/expenses/edit", h.EditPage)
	mux.HandleFunc("/expenses/update", h.UpdateExpense)
	mux.HandleFunc("/expenses/delete", h.DeleteExpense)

	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := ":8080"
	log.Printf("Сервер запущен: http://localhost%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// applyMigrations последовательно применяет миграции по номерам.
// Номер применённой хранится в PRAGMA user_version внутри файла БД.
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
	}

	return nil
}

// execFile читает SQL-файл и выполняет его.
func execFile(db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(data))
	return err
}
