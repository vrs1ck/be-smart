package models

import "time"

// Expense represents a monthly duty or bill (e.g. Rent, Electricity, Netflix).
// It belongs to a specific month/year and can be marked as covered by one person.
type Expense struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	CoveredBy string    `json:"coveredBy"` // empty string means not covered yet
	Month     int       `json:"month"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateExpenseRequest is the body the frontend sends when adding a new expense.
type CreateExpenseRequest struct {
	Title    string  `json:"title"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Month    int     `json:"month"`
	Year     int     `json:"year"`
}

// UpdateExpenseRequest uses pointer fields so the handler can tell the difference
// between a field that was not sent (nil) and a field sent as an empty value.
// This is the standard Go pattern for partial updates — only the fields you send
// will be changed; everything else stays as-is.
type UpdateExpenseRequest struct {
	Title     *string  `json:"title,omitempty"`
	Amount    *float64 `json:"amount,omitempty"`
	Category  *string  `json:"category,omitempty"`
	CoveredBy *string  `json:"coveredBy,omitempty"`
}
