package models

import "time"

// Transaction records one actual payment toward an expense in a specific month.
// Multiple transactions can exist for the same expense in the same month —
// one per person who contributed.
type Transaction struct {
	ID        int       `json:"id"`
	ExpenseID int       `json:"expenseId"`
	PaidBy    string    `json:"paidBy"`  // name of the person who paid
	Amount    float64   `json:"amount"`
	Comment   string    `json:"comment"` // optional note, e.g. "Sklep Tuesday"
	Month     int       `json:"month"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateTransactionRequest struct {
	ExpenseID int     `json:"expenseId"`
	PaidBy    string  `json:"paidBy"`
	Amount    float64 `json:"amount"`
	Comment   string  `json:"comment"`
	Month     int     `json:"month"`
	Year      int     `json:"year"`
}

type UpdateTransactionRequest struct {
	PaidBy  *string  `json:"paidBy,omitempty"`
	Amount  *float64 `json:"amount,omitempty"`
	Comment *string  `json:"comment,omitempty"`
}
