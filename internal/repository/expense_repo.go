package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"expenses/internal/models"
)

var ErrNotFound = errors.New("expense not found")

type ExpenseFilter struct {
	CategoryID int64
	From       *time.Time
	To         *time.Time
}

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

// Create сохраняет трату для конкретного пользователя.
func (r *ExpenseRepository) Create(userID int64, e *models.Expense) (int64, error) {
	var categoryID any
	if e.CategoryID > 0 {
		categoryID = e.CategoryID
	}
	res, err := r.db.Exec(
		`INSERT INTO expenses (amount, description, date, category_id, user_id) VALUES (?, ?, ?, ?, ?)`,
		e.Amount, e.Description, e.Date.Format("2006-01-02"), categoryID, userID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAll возвращает только траты конкретного пользователя.
func (r *ExpenseRepository) GetAll(userID int64, f ExpenseFilter) ([]models.Expense, error) {
	var sb strings.Builder
	sb.WriteString(`
		SELECT e.id, e.amount, e.description, e.date,
		       COALESCE(e.category_id, 0),
		       COALESCE(c.name, '')
		FROM expenses e
		LEFT JOIN categories c ON c.id = e.category_id
		WHERE e.user_id = ?
	`)

	args := []any{userID}
	var conditions []string

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
		sb.WriteString(" AND ")
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

// GetByID ищет трату этого пользователя. Не даёт посмотреть чужую.
func (r *ExpenseRepository) GetByID(userID, id int64) (*models.Expense, error) {
	var e models.Expense
	var dateStr string
	err := r.db.QueryRow(`
		SELECT e.id, e.amount, e.description, e.date,
		       COALESCE(e.category_id, 0),
		       COALESCE(c.name, '')
		FROM expenses e
		LEFT JOIN categories c ON c.id = e.category_id
		WHERE e.id = ? AND e.user_id = ?
	`, id, userID).Scan(&e.ID, &e.Amount, &e.Description, &dateStr,
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

func (r *ExpenseRepository) Update(userID int64, e *models.Expense) error {
	var categoryID any
	if e.CategoryID > 0 {
		categoryID = e.CategoryID
	}
	res, err := r.db.Exec(
		`UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ? AND user_id = ?`,
		e.Amount, e.Description, e.Date.Format("2006-01-02"), categoryID, e.ID, userID,
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

func (r *ExpenseRepository) Delete(userID, id int64) error {
	res, err := r.db.Exec(`DELETE FROM expenses WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
