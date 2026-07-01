package main

import (
	// "context"
	"encoding/json"
	"fmt"
	// "net/http"
	// "os"

	// "github.com/go-chi/chi/v5"
	// "github.com/go-chi/chi/v5/middleware"
	// "github.com/jackc/pgx/v5/pgxpool"
	// "github.com/joho/godotenv"
	"github.com/mr4tt/peppermint/models"
)

func getAccounts(accessToken string) []models.CapitalOneResp {
	url := "https://api.teller.io/accounts"

	accounts := getReq(url, accessToken)

	var accInfo []models.CapitalOneResp

	// convert response from list of json into list of CapitalOneResp type
	err := json.Unmarshal([]byte(accounts), &accInfo)
	if err != nil {
		fmt.Println("Error unmarshalling:", err)
		return nil
	}

	return accInfo
}

// get account info from a TC (teller connect) code, then use to get transactions
func getTransactions(accessToken string) []models.Transaction {
	accInfo := getAccounts(accessToken)

	var allTransactions []models.Transaction

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
		var parsedTransactions []models.Transaction
		rawTransactions := getReq(account.Links.Transactions, accessToken)
		err := json.Unmarshal((rawTransactions), &parsedTransactions)
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