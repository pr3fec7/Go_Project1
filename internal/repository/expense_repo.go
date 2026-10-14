package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"expenses/internal/models"
)

// ErrNotFound возвращается, когда запись не найдена.
var ErrNotFound = errors.New("expense not found")

// ExpenseFilter — параметры фильтрации списка трат.
// Нулевые значения означают «фильтр не задан».
type ExpenseFilter struct {
	CategoryID int64
	From       *time.Time
	To         *time.Time
}

// ExpenseRepository инкапсулирует все SQL-запросы к таблице expenses.
type ExpenseRepository struct {
	db *sql.DB
}

// NewExpenseRepository создаёт репозиторий.
func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Create вставляет новую трату и возвращает её ID.
func (r *ExpenseRepository) Create(e *models.Expense) (int64, error) {
	var categoryID any
	if e.CategoryID > 0 {
		categoryID = e.CategoryID
	}
	res, err := r.db.Exec(
		`INSERT INTO expenses (amount, description, date, category_id) VALUES (?, ?, ?, ?)`,
		e.Amount, e.Description, e.Date.Format("2006-01-02"), categoryID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAll возвращает траты по фильтру, с названием категории (JOIN).
func (r *ExpenseRepository) GetAll(f ExpenseFilter) ([]models.Expense, error) {
	var sb strings.Builder
	sb.WriteString(`
		SELECT e.id, e.amount, e.description, e.date,
		       COALESCE(e.category_id, 0),
		       COALESCE(c.name, '')
		FROM expenses e
		LEFT JOIN categories c ON c.id = e.category_id
	`)

	var conditions []string
	var args []any

	if f.CategoryID > 0 {
		conditions = append(conditions, "e.category_id = ?")
		args = append(args, f.CategoryID)
	}
	if f.From != nil {
		conditions = append(conditions, "e.date >= ?")
		args = append(args, f.From.Format("2006-01-02"))
	}
	if f.To != nil {
		conditions = append(conditions, "e.date <= ?")
		args = append(args, f.To.Format("2006-01-02"))
	}

	if len(conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conditions, " AND "))
	}

	sb.WriteString(" ORDER BY e.date DESC, e.id DESC")

	rows, err := r.db.Query(sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		var dateStr string
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &dateStr,
			&e.CategoryID, &e.CategoryName); err != nil {
			return nil, err
		}
		e.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

// GetByID возвращает трату по ID.
func (r *ExpenseRepository) GetByID(id int64) (*models.Expense, error) {
	var e models.Expense
	var dateStr string
	err := r.db.QueryRow(`
		SELECT e.id, e.amount, e.description, e.date,
		       COALESCE(e.category_id, 0),
		       COALESCE(c.name, '')
		FROM expenses e
		LEFT JOIN categories c ON c.id = e.category_id
		WHERE e.id = ?
	`, id).Scan(&e.ID, &e.Amount, &e.Description, &dateStr,
		&e.CategoryID, &e.CategoryName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	e.Date, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Update обновляет трату.
func (r *ExpenseRepository) Update(e *models.Expense) error {
	var categoryID any
	if e.CategoryID > 0 {
		categoryID = e.CategoryID
	}
	res, err := r.db.Exec(
		`UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ?`,
		e.Amount, e.Description, e.Date.Format("2006-01-02"), categoryID, e.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete удаляет трату по ID.
func (r *ExpenseRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM expenses WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
