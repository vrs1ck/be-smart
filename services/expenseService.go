package services

import (
	"fmt"
	"strings"

	"flashcards/db"
	"flashcards/models"
)

type ExpenseService struct {
	repo db.ExpenseRepository
}

func NewExpenseService(repo db.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) CreateExpense(req *models.CreateExpenseRequest) (*models.Expense, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(title) > 255 {
		return nil, fmt.Errorf("title cannot exceed 255 characters")
	}
	if req.Budget < 0 {
		return nil, fmt.Errorf("budget cannot be negative")
	}

	e := &models.Expense{
		Title:     title,
		Budget:    req.Budget,
		Recurring: req.Recurring,
	}
	if err := s.repo.CreateExpense(e); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}
	return e, nil
}

func (s *ExpenseService) GetExpenseByID(id int) (*models.Expense, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid expense ID: %d", id)
	}
	return s.repo.GetExpenseByID(id)
}

func (s *ExpenseService) GetAllExpenses() ([]*models.Expense, error) {
	return s.repo.GetAllExpenses()
}

func (s *ExpenseService) UpdateExpense(id int, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid expense ID: %d", id)
	}
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.Title == nil && req.Budget == nil && req.Recurring == nil {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	updates := make(map[string]any)

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return nil, fmt.Errorf("title cannot be empty")
		}
		updates["title"] = title
	}
	if req.Budget != nil {
		if *req.Budget < 0 {
			return nil, fmt.Errorf("budget cannot be negative")
		}
		updates["budget"] = *req.Budget
	}
	if req.Recurring != nil {
		updates["recurring"] = *req.Recurring
	}

	if err := s.repo.UpdateExpense(id, updates); err != nil {
		return nil, err
	}
	return s.repo.GetExpenseByID(id)
}

func (s *ExpenseService) DeleteExpense(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid expense ID: %d", id)
	}
	return s.repo.DeleteExpense(id)
}

// GetMonthlySummary builds the combined view for a given month:
// - All recurring expenses (always shown, even with no transactions)
// - Non-recurring expenses that have at least one transaction this month
// Each expense has its transactions for the month attached.
//
// We do two separate queries and merge in Go rather than a complex SQL join.
// This is easier to understand and maintain for a learning project.
func (s *ExpenseService) GetMonthlySummary(month, year int, txRepo db.TransactionRepository) ([]*models.ExpenseWithTransactions, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month must be between 1 and 12")
	}
	if year < 2020 {
		return nil, fmt.Errorf("year must be 2020 or later")
	}

	// Step 1: Get all recurring expenses (always in the list).
	recurring, err := s.repo.GetRecurringExpenses()
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring expenses: %w", err)
	}

	// Step 2: Get all transactions for this month.
	transactions, err := txRepo.GetTransactionsByMonth(month, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Step 3: Group transactions by expense ID using a map.
	// A map in Go is written as map[KeyType]ValueType.
	// Here: key = expense ID (int), value = slice of transactions.
	txByExpense := make(map[int][]*models.Transaction)
	for _, tx := range transactions {
		txByExpense[tx.ExpenseID] = append(txByExpense[tx.ExpenseID], tx)
	}

	// Step 4: Build the result list starting with recurring expenses.
	// We use a map to track which expense IDs are already in the list.
	seen := make(map[int]bool)
	result := make([]*models.ExpenseWithTransactions, 0)

	for _, e := range recurring {
		seen[e.ID] = true
		result = append(result, &models.ExpenseWithTransactions{
			Expense:      *e,
			Transactions: orEmpty(txByExpense[e.ID]),
		})
	}

	// Step 5: For non-recurring expenses that have transactions this month,
	// fetch the expense by ID and add it to the result.
	for expenseID, txs := range txByExpense {
		if seen[expenseID] {
			continue // already added as recurring
		}
		e, err := s.repo.GetExpenseByID(expenseID)
		if err != nil {
			continue // expense was deleted but transactions remain — skip
		}
		seen[expenseID] = true
		result = append(result, &models.ExpenseWithTransactions{
			Expense:      *e,
			Transactions: txs,
		})
	}

	return result, nil
}

// orEmpty returns the slice if non-nil, or an empty slice.
// This ensures JSON serializes as [] rather than null when there are no transactions.
func orEmpty(txs []*models.Transaction) []*models.Transaction {
	if txs == nil {
		return make([]*models.Transaction, 0)
	}
	return txs
}
