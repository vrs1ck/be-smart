package db

import (
	"database/sql"
	"fmt"

	"flashcards/models"

	_ "github.com/lib/pq"
)

// ExpenseRepository is an interface (a Go contract).
// Any type that implements all these methods satisfies this interface.
// This lets the service layer work with the database without knowing
// whether it's PostgreSQL, an in-memory store for tests, or anything else.
type ExpenseRepository interface {
	CreateExpense(expense *models.Expense) error
	GetExpenseByID(id int) (*models.Expense, error)
	GetExpensesByMonth(month, year int) ([]*models.Expense, error)
	UpdateExpense(id int, updates map[string]any) error
	DeleteExpense(id int) error
	Close() error
}

// PostgresExpenseRepository is the concrete implementation that talks to PostgreSQL.
// The db field holds the open database connection.
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

func (r *PostgresExpenseRepository) CreateExpense(expense *models.Expense) error {
	query := `
		INSERT INTO gocourse.expenses (title, amount, category, covered_by, month, year)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	// QueryRow executes the INSERT and scans the RETURNING values back into the struct.
	// This is how we get the auto-generated ID and timestamps without a second query.
	row := r.db.QueryRow(query,
		expense.Title, expense.Amount, expense.Category,
		expense.CoveredBy, expense.Month, expense.Year,
	)

	err := row.Scan(&expense.ID, &expense.CreatedAt, &expense.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create expense: %w", err)
	}

	return nil
}

func (r *PostgresExpenseRepository) GetExpenseByID(id int) (*models.Expense, error) {
	query := `
		SELECT id, title, amount, category, covered_by, month, year, created_at, updated_at
		FROM gocourse.expenses
		WHERE id = $1`

	expense := &models.Expense{}
	row := r.db.QueryRow(query, id)

	err := row.Scan(
		&expense.ID, &expense.Title, &expense.Amount, &expense.Category,
		&expense.CoveredBy, &expense.Month, &expense.Year,
		&expense.CreatedAt, &expense.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expense with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	return expense, nil
}

func (r *PostgresExpenseRepository) GetExpensesByMonth(month, year int) ([]*models.Expense, error) {
	query := `
		SELECT id, title, amount, category, covered_by, month, year, created_at, updated_at
		FROM gocourse.expenses
		WHERE month = $1 AND year = $2
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, month, year)
	if err != nil {
		return nil, fmt.Errorf("failed to query expenses: %w", err)
	}
	defer rows.Close()

	// make(slice, 0) creates an empty slice. If there are no rows, we return []
	// instead of nil — this serializes to [] in JSON rather than null.
	expenses := make([]*models.Expense, 0)
	for rows.Next() {
		expense := &models.Expense{}
		err := rows.Scan(
			&expense.ID, &expense.Title, &expense.Amount, &expense.Category,
			&expense.CoveredBy, &expense.Month, &expense.Year,
			&expense.CreatedAt, &expense.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}
		expenses = append(expenses, expense)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over expenses: %w", err)
	}

	return expenses, nil
}

// UpdateExpense builds a dynamic UPDATE statement from the map of field→value pairs.
// Only the fields present in the map are updated; everything else stays unchanged.
// The argIndex trick ($1, $2, ...) is how PostgreSQL handles positional parameters —
// unlike SQL's ? placeholders, Postgres uses numbered $N placeholders.
func (r *PostgresExpenseRepository) UpdateExpense(id int, updates map[string]any) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	query := "UPDATE gocourse.expenses SET "
	args := []any{}
	argIndex := 1

	for field, value := range updates {
		if argIndex > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	query += fmt.Sprintf(", updated_at = NOW() WHERE id = $%d", argIndex)
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("expense with id %d not found", id)
	}

	return nil
}

func (r *PostgresExpenseRepository) DeleteExpense(id int) error {
	query := "DELETE FROM gocourse.expenses WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("expense with id %d not found", id)
	}

	return nil
}

func (r *PostgresExpenseRepository) Close() error {
	return r.db.Close()
}
