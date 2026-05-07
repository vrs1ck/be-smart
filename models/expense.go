package models

import "time"

// Expense is a named spending category (e.g. "Оренда", "Корм", "Spotify").
// It does NOT belong to a specific month — it's the master list.
// Recurring=true means it shows up every month automatically.
// Recurring=false means it only appears when a transaction links to it.
type Expense struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Budget    float64   `json:"budget"`    // expected monthly amount, can be 0
	Recurring bool      `json:"recurring"` // true = appears every month
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateExpenseRequest struct {
	Title     string  `json:"title"`
	Budget    float64 `json:"budget"`
	Recurring bool    `json:"recurring"`
}

// UpdateExpenseRequest uses pointer fields so we know which fields were
// actually sent vs. omitted. A nil pointer means "don't change this field".
type UpdateExpenseRequest struct {
	Title     *string  `json:"title,omitempty"`
	Budget    *float64 `json:"budget,omitempty"`
	Recurring *bool    `json:"recurring,omitempty"`
}

// ExpenseWithTransactions is what the /monthly endpoint returns.
// The bare "Expense" line is Go struct embedding — it copies all Expense
// fields directly into this struct, then adds the Transactions slice on top.
// So in JSON you get: id, title, budget, recurring, createdAt, updatedAt, transactions.
type ExpenseWithTransactions struct {
	Expense
	Transactions []*Transaction `json:"transactions"`
}
