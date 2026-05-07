package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"flashcards/db"
	"flashcards/models"
	"flashcards/services"

	"github.com/gorilla/mux"
)

type ExpenseHandler struct {
	service *services.ExpenseService
	txRepo  db.TransactionRepository
}

// NewExpenseHandler receives the transaction repo so GetMonthlySummary
// can query both expenses and transactions.
func NewExpenseHandler(service *services.ExpenseService, txRepo db.TransactionRepository) *ExpenseHandler {
	return &ExpenseHandler{service: service, txRepo: txRepo}
}

func (h *ExpenseHandler) RegisterRoutes(router *mux.Router) {
	// /monthly must be registered before /{id} to avoid mux matching "monthly" as an ID.
	router.HandleFunc("/monthly", h.GetMonthlySummary).Methods("GET")
	router.HandleFunc("/expenses", h.CreateExpense).Methods("POST")
	router.HandleFunc("/expenses", h.GetAllExpenses).Methods("GET")
	router.HandleFunc("/expenses/{id:[0-9]+}", h.GetExpenseByID).Methods("GET")
	router.HandleFunc("/expenses/{id:[0-9]+}", h.UpdateExpense).Methods("PUT")
	router.HandleFunc("/expenses/{id:[0-9]+}", h.DeleteExpense).Methods("DELETE")
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var req models.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	expense, err := h.service.CreateExpense(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, expense)
}

// GetAllExpenses returns the master list. Supports ?recurring=true to filter.
func (h *ExpenseHandler) GetAllExpenses(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.service.GetAllExpenses()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve expenses")
		return
	}
	writeJSON(w, http.StatusOK, expenses)
}

func (h *ExpenseHandler) GetExpenseByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}
	expense, err := h.service.GetExpenseByID(id)
	if err != nil {
		if isNotFound(err.Error()) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to retrieve expense")
		}
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}
	var req models.UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	expense, err := h.service.UpdateExpense(id, &req)
	if err != nil {
		if isNotFound(err.Error()) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, expense)
}

func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}
	if err := h.service.DeleteExpense(id); err != nil {
		if isNotFound(err.Error()) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete expense")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetMonthlySummary returns recurring expenses + expenses with transactions
// for the requested month/year, with transactions nested inside each expense.
// Defaults to the current month/year if no params are given.
func (h *ExpenseHandler) GetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	q := r.URL.Query()
	if m := q.Get("month"); m != "" {
		v, err := strconv.Atoi(m)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid month parameter")
			return
		}
		month = v
	}
	if y := q.Get("year"); y != "" {
		v, err := strconv.Atoi(y)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid year parameter")
			return
		}
		year = v
	}

	summary, err := h.service.GetMonthlySummary(month, year, h.txRepo)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

// --- shared helpers (unexported, lowercase = package-private in Go) ---

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(mux.Vars(r)["id"])
}

func isNotFound(msg string) bool {
	return strings.HasSuffix(msg, "not found")
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
