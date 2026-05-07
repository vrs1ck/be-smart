package handlers

import (
	"encoding/json"
	"net/http"

	"flashcards/models"
	"flashcards/services"

	"github.com/gorilla/mux"
)

type TransactionHandler struct {
	service *services.TransactionService
}

func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/transactions", h.CreateTransaction).Methods("POST")
	router.HandleFunc("/transactions/{id:[0-9]+}", h.GetTransactionByID).Methods("GET")
	router.HandleFunc("/transactions/{id:[0-9]+}", h.UpdateTransaction).Methods("PUT")
	router.HandleFunc("/transactions/{id:[0-9]+}", h.DeleteTransaction).Methods("DELETE")
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	tx, err := h.service.CreateTransaction(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}
	tx, err := h.service.GetTransactionByID(id)
	if err != nil {
		if isNotFound(err.Error()) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to retrieve transaction")
		}
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}
	var req models.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	tx, err := h.service.UpdateTransaction(id, &req)
	if err != nil {
		if isNotFound(err.Error()) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}
	if err := h.service.DeleteTransaction(id); err != nil {
		if isNotFound(err.Error()) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete transaction")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
