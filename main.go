package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// load secrets from .env
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Error loading .env:", err)
		return
	}

	// set up auth to Teller API (SSL certs and access token)

	accessToken := os.Getenv("ACCESS_TOKEN")
	// --------------------------------

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	transactions := getTransactions(accessToken)

	r.Get("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(transactions); err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
	})

	// mounts paths from Routes() and needs them to start with /api
	r.Mount("/api", Routes())
	http.ListenAndServe("localhost:3000", r)
}

func Routes() chi.Router {
	r := chi.NewRouter()

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	fmt.Println("database url", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
	}

	handler := Handler{DBPool: pool, Token: os.Getenv("ACCESS_TOKEN")}

	// to use this, go to localhost:3000/api/...
	r.Post("/newAccount", handler.SaveUser)
	r.Get("/check/{username}", handler.CheckIfUsernameExists)

	r.Get("/{id}/remainingMoney", handler.GetRemainingMoney)
	r.Get("/{id}/updateTransactions", handler.GetNewTransactionsFromTeller)
	r.Get("/{id}/dbTransactions", handler.GetTransactionsFromDB)
	r.Get("/{id}/categories", handler.GetCategories)

	r.Post("/{id}/saveSalInfo", handler.SaveSalaryInfo)
	r.Post("/{id}/saveRecurringCosts", handler.SaveRecurringCostInfo)
	r.Post("/{id}/saveOneTimeCost", handler.SaveOneTimeCost)
	r.Post("/{id}/newCategory", handler.SaveCategories)

	r.Post("/{id}/editTransaction", handler.EditTransaction)
	r.Post("/{id}/addTransaction", handler.AddTransaction)

	r.Post("/{id}/editOneTimeCost", handler.EditOneTimeCost)
	r.Post("/{id}/editRecurringCost", handler.EditRecurringCost)

	r.Delete("/{id}/transaction", handler.DeleteTransaction)
	r.Delete("/{id}/category", handler.DeleteCategory)

	return r
}
