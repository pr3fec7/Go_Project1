package repository

import (
	"database/sql"

	"expenses/internal/models"
)

// CategoryRepository инкапсулирует SQL-запросы к таблице categories.
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository создаёт репозиторий категорий.
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// GetAll возвращает все категории по алфавиту.
func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	rows, err := r.db.Query(`SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}
