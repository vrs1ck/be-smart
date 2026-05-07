package main

import (
	"fmt"
	"log"
	"net/http"

	"flashcards/config"
	"flashcards/db"
	"flashcards/handlers"
	"flashcards/services"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	// Two repos — one connection pool per table group.
	expenseRepo, err := db.NewPostgresExpenseRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize expense database: %v", err)
	}
	defer expenseRepo.Close()

	transactionRepo, err := db.NewPostgresTransactionRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize transaction database: %v", err)
	}
	defer transactionRepo.Close()

	// Services hold business logic. Each service gets its repo as a dependency.
	expenseService := services.NewExpenseService(expenseRepo)
	transactionService := services.NewTransactionService(transactionRepo)

	// Handlers receive services (and transactionRepo for the monthly summary).
	expenseHandler := handlers.NewExpenseHandler(expenseService, transactionRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	router := mux.NewRouter()
	router.Use(corsMiddleware)
	router.Use(jsonMiddleware)

	expenseHandler.RegisterRoutes(router)
	transactionHandler.RegisterRoutes(router)
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	addr := ":" + cfg.Port
	fmt.Printf("Server starting on port %s\n", cfg.Port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}
