package repository

import (
	"database/sql"
	"errors"
	"time"

	"expenses/internal/models"
)

// возвращается ошибка, если запись не найдена.
var ErrNotFound = errors.New("expense not found")

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
	res, err := r.db.Exec(
		`INSERT INTO expenses (amount, description, date) VALUES (?, ?, ?)`,
		e.Amount, e.Description, e.Date.Format("2006-01-02"),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetAll возвращает все траты, отсортированные по дате (сначала свежие).
func (r *ExpenseRepository) GetAll() ([]models.Expense, error) {
	rows, err := r.db.Query(
		`SELECT id, amount, description, date FROM expenses ORDER BY date DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		var dateStr string
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &dateStr); err != nil {
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
	err := r.db.QueryRow(
		`SELECT id, amount, description, date FROM expenses WHERE id = ?`, id,
	).Scan(&e.ID, &e.Amount, &e.Description, &dateStr)
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
	res, err := r.db.Exec(
		`UPDATE expenses SET amount = ?, description = ?, date = ? WHERE id = ?`,
		e.Amount, e.Description, e.Date.Format("2006-01-02"), e.ID,
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
