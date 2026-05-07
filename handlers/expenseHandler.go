package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"flashcards/models"
	"flashcards/services"

	"github.com/gorilla/mux"
)

type ExpenseHandler struct {
	service *services.ExpenseService
}

func NewExpenseHandler(service *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

func (h *ExpenseHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/expenses", h.CreateExpense).Methods("POST")
	router.HandleFunc("/expenses", h.GetExpensesByMonth).Methods("GET")
	router.HandleFunc("/expenses/{id:[0-9]+}", h.GetExpenseByID).Methods("GET")
	router.HandleFunc("/expenses/{id:[0-9]+}", h.UpdateExpense).Methods("PUT")
	router.HandleFunc("/expenses/{id:[0-9]+}", h.DeleteExpense).Methods("DELETE")
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var req models.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	expense, err := h.service.CreateExpense(&req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, expense)
}

// GetExpensesByMonth reads ?month= and ?year= from the URL query string.
// If not provided, it defaults to the current month and year.
// This way the frontend can call GET /expenses with no params and always get "this month".
func (h *ExpenseHandler) GetExpensesByMonth(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	month := now.Month()
	year := now.Year()

	query := r.URL.Query()

	if m := query.Get("month"); m != "" {
		parsed, err := strconv.Atoi(m)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid month parameter")
			return
		}
		month = time.Month(parsed)
	}

	if y := query.Get("year"); y != "" {
		parsed, err := strconv.Atoi(y)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid year parameter")
			return
		}
		year = parsed
	}

	expenses, err := h.service.GetExpensesByMonth(int(month), year)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, expenses)
}

func (h *ExpenseHandler) GetExpenseByID(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	expense, err := h.service.GetExpenseByID(id)
	if err != nil {
		if isNotFound(err.Error()) {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve expense")
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, expense)
}

func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	var req models.UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	expense, err := h.service.UpdateExpense(id, &req)
	if err != nil {
		if isNotFound(err.Error()) {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.writeJSONResponse(w, http.StatusOK, expense)
}

func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	id, err := h.parseID(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	if err := h.service.DeleteExpense(id); err != nil {
		if isNotFound(err.Error()) {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete expense")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ExpenseHandler) parseID(r *http.Request) (int, error) {
	return strconv.Atoi(mux.Vars(r)["id"])
}

func (h *ExpenseHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *ExpenseHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func isNotFound(message string) bool {
	return strings.HasSuffix(message, "not found")
}
