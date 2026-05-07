package services

import (
	"fmt"
	"strings"

	"flashcards/db"
	"flashcards/models"
)

// ExpenseService sits between the handler (HTTP) and the repository (database).
// Its job is: validate inputs, apply business rules, then delegate to the repo.
// The handler never talks to the database directly — it always goes through here.
type ExpenseService struct {
	repo db.ExpenseRepository
}

// NewExpenseService takes an ExpenseRepository interface, not the concrete PostgreSQL type.
// This is Go's dependency injection: the service works with any repo implementation.
func NewExpenseService(repo db.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) CreateExpense(req *models.CreateExpenseRequest) (*models.Expense, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	expense := &models.Expense{
		Title:     strings.TrimSpace(req.Title),
		Amount:    req.Amount,
		Category:  strings.TrimSpace(req.Category),
		CoveredBy: "",
		Month:     req.Month,
		Year:      req.Year,
	}

	if err := s.repo.CreateExpense(expense); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return expense, nil
}

func (s *ExpenseService) GetExpenseByID(id int) (*models.Expense, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid expense ID: %d", id)
	}
	return s.repo.GetExpenseByID(id)
}

func (s *ExpenseService) GetExpensesByMonth(month, year int) ([]*models.Expense, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month must be between 1 and 12")
	}
	if year < 2020 {
		return nil, fmt.Errorf("year must be 2020 or later")
	}
	return s.repo.GetExpensesByMonth(month, year)
}

func (s *ExpenseService) UpdateExpense(id int, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid expense ID: %d", id)
	}

	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Build a map of only the fields that were actually sent.
	// The DB layer uses this map to build a dynamic UPDATE statement.
	updates := make(map[string]any)

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return nil, fmt.Errorf("title cannot be empty")
		}
		updates["title"] = trimmed
	}

	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}

	if req.Category != nil {
		updates["category"] = strings.TrimSpace(*req.Category)
	}

	if req.CoveredBy != nil {
		updates["covered_by"] = strings.TrimSpace(*req.CoveredBy)
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no valid updates provided")
	}

	if err := s.repo.UpdateExpense(id, updates); err != nil {
		return nil, err
	}

	// Fetch and return the updated record so the caller gets fresh data.
	return s.repo.GetExpenseByID(id)
}

func (s *ExpenseService) DeleteExpense(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid expense ID: %d", id)
	}
	return s.repo.DeleteExpense(id)
}

func (s *ExpenseService) validateCreateRequest(req *models.CreateExpenseRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return fmt.Errorf("title is required")
	}
	if len(title) > 255 {
		return fmt.Errorf("title cannot exceed 255 characters")
	}

	if req.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	if len(strings.TrimSpace(req.Category)) > 50 {
		return fmt.Errorf("category cannot exceed 50 characters")
	}

	if req.Month < 1 || req.Month > 12 {
		return fmt.Errorf("month must be between 1 and 12")
	}

	if req.Year < 2020 {
		return fmt.Errorf("year must be 2020 or later")
	}

	return nil
}

func (s *ExpenseService) validateUpdateRequest(req *models.UpdateExpenseRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.Title == nil && req.Amount == nil && req.Category == nil && req.CoveredBy == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}

	if req.Amount != nil && *req.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	if req.Category != nil && len(strings.TrimSpace(*req.Category)) > 50 {
		return fmt.Errorf("category cannot exceed 50 characters")
	}

	if req.CoveredBy != nil && len(strings.TrimSpace(*req.CoveredBy)) > 50 {
		return fmt.Errorf("covered_by cannot exceed 50 characters")
	}

	return nil
}
