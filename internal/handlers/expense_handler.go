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
	expenses   *repository.ExpenseRepository
	categories *repository.CategoryRepository
	indexTmpl  *template.Template
	addTmpl    *template.Template
	editTmpl   *template.Template
}

func NewExpenseHandler(
	expenses *repository.ExpenseRepository,
	categories *repository.CategoryRepository,
) *ExpenseHandler {
	layout := filepath.Join("web", "templates", "layout.html")
	index := filepath.Join("web", "templates", "index.html")
	add := filepath.Join("web", "templates", "add.html")
	edit := filepath.Join("web", "templates", "edit.html")

	return &ExpenseHandler{
		expenses:   expenses,
		categories: categories,
		indexTmpl:  template.Must(template.ParseFiles(layout, index)),
		addTmpl:    template.Must(template.ParseFiles(layout, add)),
		editTmpl:   template.Must(template.ParseFiles(layout, edit)),
	}
}

// ListPage отображает список трат с учётом фильтров.
func (h *ExpenseHandler) ListPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	filter := parseFilter(r)

	expenses, err := h.expenses.GetAll(filter)
	if err != nil {
		log.Printf("GetAll: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	cats, err := h.categories.GetAll()
	if err != nil {
		log.Printf("CategoryRepo.GetAll: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":            "Список трат",
		"Expenses":         expenses,
		"Categories":       cats,
		"SelectedCategory": filter.CategoryID,
		"From":             r.URL.Query().Get("from"),
		"To":               r.URL.Query().Get("to"),
	}
	if err := h.indexTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("index template: %v", err)
	}
}

func (h *ExpenseHandler) NewExpensePage(w http.ResponseWriter, r *http.Request) {
	cats, err := h.categories.GetAll()
	if err != nil {
		log.Printf("CategoryRepo.GetAll: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":      "Добавить трату",
		"Action":     "/expenses",
		"Categories": cats,
		"CategoryID": int64(0),
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

	var categoryID int64
	if catStr := r.FormValue("category_id"); catStr != "" {
		categoryID, _ = strconv.ParseInt(catStr, 10, 64)
	}

	_, err = h.expenses.Create(&models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
		CategoryID:  categoryID,
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

	exp, err := h.expenses.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("GetByID: %v", err)
		http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	cats, err := h.categories.GetAll()
	if err != nil {
		log.Printf("CategoryRepo.GetAll: %v", err)
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
		"Categories":  cats,
		"CategoryID":  exp.CategoryID,
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

	var categoryID int64
	if catStr := r.FormValue("category_id"); catStr != "" {
		categoryID, _ = strconv.ParseInt(catStr, 10, 64)
	}

	err = h.expenses.Update(&models.Expense{
		ID:          id,
		Amount:      amount,
		Description: description,
		Date:        date,
		CategoryID:  categoryID,
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

	err = h.expenses.Delete(id)
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

// parseFilter читает query-параметры category, from, to.
func parseFilter(r *http.Request) repository.ExpenseFilter {
	var f repository.ExpenseFilter

	if cat := r.URL.Query().Get("category"); cat != "" {
		if id, err := strconv.ParseInt(cat, 10, 64); err == nil {
			f.CategoryID = id
		}
	}
	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			f.From = &t
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			f.To = &t
		}
	}
	return f
}
