package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr4tt/peppermint/models"
)

type Handler struct {
	DBPool *pgxpool.Pool
}

func (b Handler) SaveUser(w http.ResponseWriter, r *http.Request) {
	var userInfo models.User
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Println("info received: ", userInfo)

	userQuery := `INSERT INTO users (username, pw_hash) VALUES (@name, @hash)`
	args := pgx.NamedArgs{
		"name": userInfo.Username,
		"hash": userInfo.Password,
	}

	_, err := b.DBPool.Exec(context.Background(), userQuery, args)

	if err != nil {
		fmt.Printf("SaveUser insert query failed: %v\n", err)
		return
	}
}

func (b Handler) SaveSalaryInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var userInfo models.SalaryInfo
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Println("info received: ", userInfo)

	salaryInfoQuery := `
	INSERT INTO user_finances (user_id, amt_401k_contribution, total_insurance_amount, monthly_posttax_salary)
	VALUES (@uid, @k401, @insurance, @salary)`

	args := pgx.NamedArgs{
		"uid":       id,
		"k401":      userInfo.K401,
		"insurance": userInfo.Insurance,
		"salary":    userInfo.PostTaxSal,
	}

	_, err := b.DBPool.Exec(context.Background(), salaryInfoQuery, args)
	if err != nil {
		fmt.Printf("SalaryInfo insert query failed: %v\n", err)
		return
	}
}

func (b Handler) SaveRecurringCostInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var recurringCosts []models.RecurringCost
	if err := json.NewDecoder(r.Body).Decode(&recurringCosts); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	recurringCostQuery := `
	INSERT INTO recurring_costs (user_id, name, amount, month_frequency, is_savings)
	VALUES (@uid, @name, @amt, @freq, @isSavings)
	`

	for _, userCosts := range recurringCosts {
		args := pgx.NamedArgs{
			"uid":       id,
			"name":      userCosts.Name,
			"amt":       userCosts.Amount,
			"freq":      userCosts.Frequency,
			"isSavings": userCosts.IsSavings,
		}
		_, err := b.DBPool.Exec(context.Background(), recurringCostQuery, args)

		if err != nil {
			fmt.Printf("recurring cost insert failed: %v\n", err)
			return
		}
	}
}

func (b Handler) SaveOneTimeCost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var userCosts []models.OneTimeCost
	if err := json.NewDecoder(r.Body).Decode(&userCosts); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Println("received: ", userCosts)

	oneTimeCostQuery := `
	INSERT INTO onetime_costs (user_id, name, amount, month, year)
	VALUES (@uid, @name, @amt, @month, @year)
	`

	for _, oneTimeCost := range userCosts {
		args := pgx.NamedArgs{
			"uid":   id,
			"name":  oneTimeCost.Name,
			"amt":   oneTimeCost.Amount,
			"month": oneTimeCost.Month,
			"year":  oneTimeCost.Year,
		}

		_, err := b.DBPool.Exec(context.Background(), oneTimeCostQuery, args)

		if err != nil {
			fmt.Printf("one time cost query insert failed: %v\n", err)
			return
		}
	}
}

func (b Handler) SaveCategories(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var userCategories []models.CategoryInfo
	if err := json.NewDecoder(r.Body).Decode(&userCategories); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Println("receieved: ", userCategories)

	categoryQuery := `INSERT INTO saved_categories (user_id, category_name, category_limit) 
	VALUES (@uid, @name, @lim)`

	for _, category := range userCategories {
		args := pgx.NamedArgs{
			"uid":  id,
			"name": category.Name,
			"lim":  category.Limit,
		}

		_, err := b.DBPool.Exec(context.Background(), categoryQuery, args)

		if err != nil {
			fmt.Printf("save categories insert failed: %v\n", err)
			return
		}
	}
}

func (b Handler) AddTransaction(w http.ResponseWriter, r *http.Request)  {}
func (b Handler) EditTransaction(w http.ResponseWriter, r *http.Request) {}

// for updating costs where name / amount / whatever needs to change
func (b Handler) EditOneTimeCost(w http.ResponseWriter, r *http.Request)   {}
func (b Handler) EditRecurringCost(w http.ResponseWriter, r *http.Request) {}

func (b Handler) GetRemainingMoney(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	w.Write([]byte(id))
}

func (b Handler) GetNewTransactionsFromTeller(w http.ResponseWriter, r *http.Request) {}
func (b Handler) GetTransactionsFromDB(w http.ResponseWriter, r *http.Request)        {}

func (b Handler) GetCategories(w http.ResponseWriter, r *http.Request) {}

func (b Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {}
func (b Handler) DeleteCategory(w http.ResponseWriter, r *http.Request)    {}
