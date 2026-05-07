package db

import (
	"database/sql"
	"fmt"

	"flashcards/models"

	_ "github.com/lib/pq"
)

type TransactionRepository interface {
	CreateTransaction(t *models.Transaction) error
	GetTransactionByID(id int) (*models.Transaction, error)
	GetTransactionsByMonth(month, year int) ([]*models.Transaction, error)
	GetTransactionsByExpenseAndMonth(expenseID, month, year int) ([]*models.Transaction, error)
	UpdateTransaction(id int, updates map[string]any) error
	DeleteTransaction(id int) error
	Close() error
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(databaseURL string) (*PostgresTransactionRepository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &PostgresTransactionRepository{db: db}, nil
}

func (r *PostgresTransactionRepository) CreateTransaction(t *models.Transaction) error {
	query := `
		INSERT INTO gocourse.transactions (expense_id, paid_by, amount, comment, month, year)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.db.QueryRow(query,
		t.ExpenseID, t.PaidBy, t.Amount, t.Comment, t.Month, t.Year,
	).Scan(&t.ID, &t.CreatedAt)
}

func (r *PostgresTransactionRepository) GetTransactionByID(id int) (*models.Transaction, error) {
	query := `
		SELECT id, expense_id, paid_by, amount, comment, month, year, created_at
		FROM gocourse.transactions WHERE id = $1`

	t := &models.Transaction{}
	err := r.db.QueryRow(query, id).Scan(
		&t.ID, &t.ExpenseID, &t.PaidBy, &t.Amount, &t.Comment, &t.Month, &t.Year, &t.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return t, nil
}

func (r *PostgresTransactionRepository) GetTransactionsByMonth(month, year int) ([]*models.Transaction, error) {
	query := `
		SELECT id, expense_id, paid_by, amount, comment, month, year, created_at
		FROM gocourse.transactions
		WHERE month = $1 AND year = $2
		ORDER BY expense_id, created_at ASC`

	return r.scanTransactions(r.db.Query(query, month, year))
}

func (r *PostgresTransactionRepository) GetTransactionsByExpenseAndMonth(expenseID, month, year int) ([]*models.Transaction, error) {
	query := `
		SELECT id, expense_id, paid_by, amount, comment, month, year, created_at
		FROM gocourse.transactions
		WHERE expense_id = $1 AND month = $2 AND year = $3
		ORDER BY created_at ASC`

	return r.scanTransactions(r.db.Query(query, expenseID, month, year))
}

func (r *PostgresTransactionRepository) scanTransactions(rows *sql.Rows, err error) ([]*models.Transaction, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*models.Transaction, 0)
	for rows.Next() {
		t := &models.Transaction{}
		if err := rows.Scan(&t.ID, &t.ExpenseID, &t.PaidBy, &t.Amount, &t.Comment, &t.Month, &t.Year, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

func (r *PostgresTransactionRepository) UpdateTransaction(id int, updates map[string]any) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	query := "UPDATE gocourse.transactions SET "
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
	query += fmt.Sprintf(" WHERE id = $%d", i)
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("transaction with id %d not found", id)
	}
	return nil
}

func (r *PostgresTransactionRepository) DeleteTransaction(id int) error {
	result, err := r.db.Exec("DELETE FROM gocourse.transactions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("transaction with id %d not found", id)
	}
	return nil
}

func (r *PostgresTransactionRepository) Close() error {
	return r.db.Close()
}
