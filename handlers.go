package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (Pool = func() *pgxpool.Pool {
		dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		}
		// defer dbpool.Close()

		return dbpool
	}()
)
type Handler struct {
	DBPool *pgxpool.Pool
}

// save user’s post tax salary
// save user’s 401k contribution / insurance / other pre-tax stuff
// save user’s % going to saving things (this could be, say, a car, savings account, stocks, etc)
// save user’s utility/fixed-cost bills (e.g., electricity, rent, insurance, subscriptions)
// save category + how much money user allocates to that category (the actual budgeting stuff lol) + notes for that category

type SalaryInfo struct {
    K401      float64 `json:"contribution_401"`
    Insurance float64 `json:"total_insurance_amount"`
    PostTaxSal float64 `json:"monthly_posttax_salary"`
}
type OneTimeCost struct {
    Name   string `json:"name"`
    Amount int64  `json:"amount"`
    Month  int    `json:"month"`
    Year   int    `json:"year"`
}

type RecurringCost struct {
    Name      string  `json:"name"`
    Amount    float64 `json:"amount"`
    Frequency int64   `json:"frequency"`
    IsSavings bool    `json:"is_savings"`
}

type CategoryInfo struct {
	
}

// saves user information 
func (b Handler) SaveSalaryInfo(w http.ResponseWriter, r *http.Request)    {
	id := chi.URLParam(r, "id")

	var userInfo SalaryInfo
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	salaryInfoQuery := `
	INSERT INTO UserFinances (user_id, amt_401k_contribution, total_insurance_amount, monthly_posttax_salary)
	VALUES (?, ?, ?, ?)
	`
	_, err := b.DBPool.Exec(context.Background(), salaryInfoQuery, 
	id, userInfo.K401, userInfo.Insurance, userInfo.PostTaxSal)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return
	}
}

func (b Handler) SaveRecurringCostInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var userCosts []RecurringCost
	if err := json.NewDecoder(r.Body).Decode(&userCosts); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	recurringCostQuery := `
	INSERT INTO RecurringCosts (user_id, name, amount, month_frequency, is_savings)
	VALUES (?, ?, ?, ?, ?)
	`

	for _, recurringCost := range userCosts { 
		_, err := b.DBPool.Exec(context.Background(), recurringCostQuery, 
		id, recurringCost.Name, recurringCost.Amount, recurringCost.Frequency, recurringCost.IsSavings)

		if err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
			return
		}
	}
}

func (b Handler) SaveOneTimeCost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var userCosts []OneTimeCost
	if err := json.NewDecoder(r.Body).Decode(&userCosts); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	oneTimeCostQuery := `
	INSERT INTO OneTimeCosts (user_id, name, amount, month, year)
	VALUES (?, ?, ?, ?, ?)
	`

	for _, oneTimeCost := range userCosts { 
		_, err := b.DBPool.Exec(context.Background(), oneTimeCostQuery, 
		id, oneTimeCost.Name, oneTimeCost.Amount, oneTimeCost.Month, oneTimeCost.Year)

		if err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
			return
		}
	}
}

func (b Handler) SaveCategories(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var userCosts []OneTimeCost
	if err := json.NewDecoder(r.Body).Decode(&userCosts); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	oneTimeCostQuery := `
	INSERT INTO OneTimeCosts (user_id, name, amount, month, year)
	VALUES (?, ?, ?, ?, ?)
	`

	for _, oneTimeCost := range userCosts { 
		_, err := b.DBPool.Exec(context.Background(), oneTimeCostQuery, 
		id, oneTimeCost.Name, oneTimeCost.Amount, oneTimeCost.Month, oneTimeCost.Year)

		if err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
			return
		}
	}
}

func (b Handler) SaveNewCategory(w http.ResponseWriter, r *http.Request) {}

func (b Handler) AddTransaction(w http.ResponseWriter, r *http.Request)  {}
func (b Handler) EditTransaction(w http.ResponseWriter, r *http.Request) {}

func (b Handler) GetRemainingMoney(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	w.Write([]byte(id))
}
func (b Handler) GetNewTransactionsFromTeller(w http.ResponseWriter, r *http.Request) {}
func (b Handler) GetTransactionsFromDB(w http.ResponseWriter, r *http.Request)        {}

func (b Handler) GetCategories(w http.ResponseWriter, r *http.Request) {}

func (b Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {}
func (b Handler) DeleteCategory(w http.ResponseWriter, r *http.Request)    {}
