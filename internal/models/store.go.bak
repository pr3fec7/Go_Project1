package models

type ExpenseStore struct {
	expenses []Expense
	nextID   int
}

func NewExpenseStore() *ExpenseStore {
	return &ExpenseStore{
		expenses: []Expense{},
		nextID:   1,
	}
}

func (s *ExpenseStore) Add(e Expense) Expense {
	e.ID = s.nextID
	s.nextID++
	s.expenses = append(s.expenses, e)
	return e
}

func (s *ExpenseStore) All() []Expense {
	return s.expenses
}
