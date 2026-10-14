CREATE TABLE IF NOT EXISTS categories (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT    NOT NULL UNIQUE
);

INSERT OR IGNORE INTO categories (name) VALUES
    ('Еда'),
    ('Транспорт'),
    ('Жильё'),
    ('Развлечения'),
    ('Здоровье'),
    ('Другое');

ALTER TABLE expenses ADD COLUMN category_id INTEGER REFERENCES categories(id);

CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category_id);