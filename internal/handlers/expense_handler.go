package handlers

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"expenses/internal/models"
)

type ExpenseHandler struct {
	store     *models.ExpenseStore
	indexTmpl *template.Template
	addTmpl   *template.Template
}

func NewExpenseHandler(store *models.ExpenseStore) *ExpenseHandler {
	indexTmpl := template.Must(template.ParseFiles(
		filepath.Join("web", "templates", "layout.html"),
		filepath.Join("web", "templates", "index.html"),
	))
	addTmpl := template.Must(template.ParseFiles(
		filepath.Join("web", "templates", "layout.html"),
		filepath.Join("web", "templates", "add.html"),
	))
	return &ExpenseHandler{
		store:     store,
		indexTmpl: indexTmpl,
		addTmpl:   addTmpl,
	}
}

func (h *ExpenseHandler) ListPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]any{
		"Title":    "Список трат",
		"Expenses": h.store.All(),
	}

	if err := h.indexTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ошибка рендера шаблона: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
	}
}

func (h *ExpenseHandler) NewExpensePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Добавить трату",
	}

	if err := h.addTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ошибка рендера шаблона: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
	}
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	amountStr := r.FormValue("amount")
	description := r.FormValue("description")
	dateStr := r.FormValue("date")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Некорректная дата", http.StatusBadRequest)
		return
	}

	if amount <= 0 {
		http.Error(w, "Сумма должна быть положительной", http.StatusBadRequest)
		return
	}

	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}

	h.store.Add(models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
