package main

import (
	"log"
	"net/http"

	"expenses/internal/handlers"
	"expenses/internal/models"
)

func main() {
	store := models.NewExpenseStore()
	h := handlers.NewExpenseHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("/", h.ListPage)
	mux.HandleFunc("/expenses/new", h.NewExpensePage)
	mux.HandleFunc("/expenses", h.CreateExpense)

	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := ":8080"
	log.Printf("Сервер запущен: http://localhost%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
