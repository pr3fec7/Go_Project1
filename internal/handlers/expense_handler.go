package handlers

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"expenses/internal/models"
	"expenses/internal/repository"
)

type ExpenseHandler struct {
	repo      *repository.ExpenseRepository
	indexTmpl *template.Template
	addTmpl   *template.Template
	editTmpl  *template.Template
}

func NewExpenseHandler(repo *repository.ExpenseRepository) *ExpenseHandler {
	layout := filepath.Join("web", "templates", "layout.html")
	index := filepath.Join("web", "templates", "index.html")
	add := filepath.Join("web", "templates", "add.html")
	edit := filepath.Join("web", "templates", "edit.html")

	return &ExpenseHandler{
		repo:      repo,
		indexTmpl: template.Must(template.ParseFiles(layout, index)),
		addTmpl:   template.Must(template.ParseFiles(layout, add)),
		editTmpl:  template.Must(template.ParseFiles(layout, edit)),
	}
}

func (h *ExpenseHandler) ListPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	expenses, err := h.repo.GetAll()
	if err != nil {
		log.Printf("GetAll: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":    "Список трат",
		"Expenses": expenses,
	}
	if err := h.indexTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("index template: %v", err)
	}
}

func (h *ExpenseHandler) NewExpensePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":  "Добавить трату",
		"Action": "/expenses",
	}
	if err := h.addTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("add template: %v", err)
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

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "Описание не может быть пустым", http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		http.Error(w, "Некорректная дата", http.StatusBadRequest)
		return
	}

	_, err = h.repo.Create(&models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
	})
	if err != nil {
		log.Printf("Create: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *ExpenseHandler) EditPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	exp, err := h.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("GetByID: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":       "Редактировать трату",
		"Action":      "/expenses/update",
		"ID":          exp.ID,
		"Amount":      exp.Amount,
		"Description": exp.Description,
		"Date":        exp.Date.Format("2006-01-02"),
	}
	if err := h.editTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("edit template: %v", err)
	}
}

func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Некорректная сумма", http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "Описание не может быть пустым", http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		http.Error(w, "Некорректная дата", http.StatusBadRequest)
		return
	}

	err = h.repo.Update(&models.Expense{
		ID:          id,
		Amount:      amount,
		Description: description,
		Date:        date,
	})
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("Update: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(id)
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("Delete: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
