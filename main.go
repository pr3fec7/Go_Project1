package main

import (
	"database/sql"
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

	repo := repository.NewExpenseRepository(db)
	h := handlers.NewExpenseHandler(repo)

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

// applyMigrations читает SQL-файл и выполняет его.
func applyMigrations(db *sql.DB) error {
	data, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(data))
	return err
}
