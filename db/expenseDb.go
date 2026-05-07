package db

import (
	"database/sql"
	"fmt"

	"flashcards/models"

	_ "github.com/lib/pq"
)

type ExpenseRepository interface {
	CreateExpense(e *models.Expense) error
	GetExpenseByID(id int) (*models.Expense, error)
	GetAllExpenses() ([]*models.Expense, error)
	GetRecurringExpenses() ([]*models.Expense, error)
	UpdateExpense(id int, updates map[string]any) error
	DeleteExpense(id int) error
	Close() error
}

type PostgresExpenseRepository struct {
	db *sql.DB
}

func NewPostgresExpenseRepository(databaseURL string) (*PostgresExpenseRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &PostgresExpenseRepository{db: db}, nil
}

func (r *PostgresExpenseRepository) CreateExpense(e *models.Expense) error {
	query := `
		INSERT INTO gocourse.expenses (title, budget, recurring)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRow(query, e.Title, e.Budget, e.Recurring)
	return row.Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *PostgresExpenseRepository) GetExpenseByID(id int) (*models.Expense, error) {
	query := `
		SELECT id, title, budget, recurring, created_at, updated_at
		FROM gocourse.expenses WHERE id = $1`

	e := &models.Expense{}
	err := r.db.QueryRow(query, id).Scan(
		&e.ID, &e.Title, &e.Budget, &e.Recurring, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("expense with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}
	return e, nil
}

func (r *PostgresExpenseRepository) GetAllExpenses() ([]*models.Expense, error) {
	return r.queryExpenses(`
		SELECT id, title, budget, recurring, created_at, updated_at
		FROM gocourse.expenses
		ORDER BY title ASC`)
}

func (r *PostgresExpenseRepository) GetRecurringExpenses() ([]*models.Expense, error) {
	return r.queryExpenses(`
		SELECT id, title, budget, recurring, created_at, updated_at
		FROM gocourse.expenses
		WHERE recurring = true
		ORDER BY title ASC`)
}

// queryExpenses is a shared helper so GetAllExpenses and GetRecurringExpenses
// don't duplicate the scan logic. In Go, private helpers start with a lowercase letter.
func (r *PostgresExpenseRepository) queryExpenses(query string, args ...any) ([]*models.Expense, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]*models.Expense, 0)
	for rows.Next() {
		e := &models.Expense{}
		if err := rows.Scan(&e.ID, &e.Title, &e.Budget, &e.Recurring, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

func (r *PostgresExpenseRepository) UpdateExpense(id int, updates map[string]any) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	query := "UPDATE gocourse.expenses SET "
	args := []any{}
	i := 1
	for field, value := range updates {
		if i > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, i)
		args = append(args, value)
		i++
	}
	query += fmt.Sprintf(", updated_at = NOW() WHERE id = $%d", i)
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update expense: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("expense with id %d not found", id)
	}
	return nil
}

func (r *PostgresExpenseRepository) DeleteExpense(id int) error {
	result, err := r.db.Exec("DELETE FROM gocourse.expenses WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("expense with id %d not found", id)
	}
	return nil
}

func (r *PostgresExpenseRepository) Close() error {
	return r.db.Close()
}
