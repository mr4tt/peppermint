package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

var (
	TClient = func(certFile string, keyFile string) *http.Client {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			fmt.Println("Error loading certificates:", err)
			return nil
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		return &http.Client{Transport: transport}
	}("certs/certificate.pem", "certs/private_key.pem")
)

// make a GET request, given auth, url to request, and a client (for certs)
func getReq(url string, accessToken string) []byte {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating new HTTP request:", err)
		return nil
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(accessToken, "")

	// make the http request
	response, err := TClient.Do(request)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}
	defer response.Body.Close()

	fullResponse, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	return fullResponse
}

// get account info from a TC (teller connect) code, then use to get transactions
func getTransactions(accessToken string) []Transaction {
	url := "https://api.teller.io/accounts"

	accounts := getReq(url, accessToken)

	var accInfo []CapitalOneResp

	// convert response from list of json into list of CapitalOneResp type
	err := json.Unmarshal([]byte(accounts), &accInfo)
	if err != nil {
		fmt.Println("Error unmarshalling:", err)
		return nil
	}

	var allTransactions []Transaction

	// subtypes of accounts are
	// depository:
	// checking, savings, money_market, certificate_of_deposit, treasury, sweep
	// credit:
	// credit_card

	// for each account found, get the transactions from it and
	// convert to Transaction type
	for _, account := range accInfo {
		if account.Subtype != "checking" && account.Subtype != "credit_card" {
			continue
		}

		fmt.Println("ID:", account.ID)
		fmt.Println("Name:", account.Name)

		// Get all transactions associated with this account
		var parsedTransactions []Transaction
		rawTransactions := getReq(account.Links.Transactions, accessToken)
		err = json.Unmarshal((rawTransactions), &parsedTransactions)
		if err != nil {
			fmt.Println("Error unmarshalling transactions:", err)
			return nil
		}

		for _, transaction := range parsedTransactions {
			// We only want to process posted transactions
			if transaction.Status != "posted" {
				continue
			}

			allTransactions = append(allTransactions, transaction)
		}
	}

	return allTransactions
}

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

	http.ListenAndServe(":3000", r)

}

func Routes() chi.Router {
	r := chi.NewRouter()

	handler := Handler{}

	// to use this, go to localhost:3000/api/{id}/moneyLeft
	r.Get("/{id}/moneyLeft", handler.GetRemainingMoney)
	r.Get("/{id}/tellerTransactions", handler.GetNewTransactionsFromTeller)
	r.Get("/{id}/dbTransactions", handler.GetTransactionsFromDB)
	r.Get("/{id}/categories", handler.GetCategories)

	r.Post("/{id}/saveInfo", handler.SaveUserInfo)
	r.Post("/{id}/editTransaction", handler.EditTransaction)
	r.Post("/{id}/addTransaction", handler.AddTransaction)
	r.Post("/{id}/newCategory", handler.SaveNewCategory)

	r.Delete("/{id}/transaction", handler.DeleteTransaction)
	r.Delete("/{id}/category", handler.DeleteCategory)

	return r
}
