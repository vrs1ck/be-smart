package services

import (
	"fmt"
	"strings"

	"flashcards/db"
	"flashcards/models"
)

type TransactionService struct {
	repo db.TransactionRepository
}

func NewTransactionService(repo db.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) CreateTransaction(req *models.CreateTransactionRequest) (*models.Transaction, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.ExpenseID <= 0 {
		return nil, fmt.Errorf("expenseId is required")
	}

	paidBy := strings.TrimSpace(req.PaidBy)
	if paidBy == "" {
		return nil, fmt.Errorf("paidBy is required")
	}
	if len(paidBy) > 100 {
		return nil, fmt.Errorf("paidBy cannot exceed 100 characters")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	if req.Month < 1 || req.Month > 12 {
		return nil, fmt.Errorf("month must be between 1 and 12")
	}
	if req.Year < 2020 {
		return nil, fmt.Errorf("year must be 2020 or later")
	}

	t := &models.Transaction{
		ExpenseID: req.ExpenseID,
		PaidBy:    paidBy,
		Amount:    req.Amount,
		Comment:   strings.TrimSpace(req.Comment),
		Month:     req.Month,
		Year:      req.Year,
	}
	if err := s.repo.CreateTransaction(t); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	return t, nil
}

func (s *TransactionService) GetTransactionByID(id int) (*models.Transaction, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid transaction ID: %d", id)
	}
	return s.repo.GetTransactionByID(id)
}

func (s *TransactionService) UpdateTransaction(id int, req *models.UpdateTransactionRequest) (*models.Transaction, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid transaction ID: %d", id)
	}
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.PaidBy == nil && req.Amount == nil && req.Comment == nil {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	updates := make(map[string]any)

	if req.PaidBy != nil {
		paidBy := strings.TrimSpace(*req.PaidBy)
		if paidBy == "" {
			return nil, fmt.Errorf("paidBy cannot be empty")
		}
		updates["paid_by"] = paidBy
	}
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return nil, fmt.Errorf("amount must be greater than 0")
		}
		updates["amount"] = *req.Amount
	}
	if req.Comment != nil {
		updates["comment"] = strings.TrimSpace(*req.Comment)
	}

	if err := s.repo.UpdateTransaction(id, updates); err != nil {
		return nil, err
	}
	return s.repo.GetTransactionByID(id)
}

func (s *TransactionService) DeleteTransaction(id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid transaction ID: %d", id)
	}
	return s.repo.DeleteTransaction(id)
}
