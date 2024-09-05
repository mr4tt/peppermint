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

// response: https://teller.io/docs/api/identity#get-identity
// links: https://teller.io/docs/api/accounts#properties
type CapitalOneResp struct {
	Type         string `json:"type"`
	Subtype      string `json:"subtype"`
	Status       string `json:"status"`
	Name         string `json:"name"`
	Links        struct {
		Transactions string `json:"transactions"`
		Self         string `json:"self"`
		Balances     string `json:"balances"`
	} `json:"links"`
	Institution struct {
		Name         string `json:"name"`
		ID           string `json:"id"`
	} `json:"institution"`
	ID           string `json:"id"`
	EnrollmentID string `json:"enrollment_id"`
}

// make a GET request, given auth, url to request, and a client (for certs)
func getReq(url string, accessToken string, client *http.Client) []byte {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating new HTTP request:", err)
		return nil
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(accessToken, "")

	// make the http request
	response, err := client.Do(request)
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

// get account info from TC (teller connect) code, then use to get transactions
func getTransactions(accessToken string, client *http.Client) [][]byte {
	url := "https://api.teller.io/accounts"

	accounts := getReq(url, accessToken, client) 

	var accInfo []CapitalOneResp

	err := json.Unmarshal([]byte(accounts), &accInfo) 
	if err != nil {
		fmt.Println("Error unmarshalling:", err)
		return nil
	}
	var transactions [][]byte 
	for _, account := range accInfo {
		fmt.Println("ID:", account.ID)
		fmt.Println("Name:", account.Name)

		// should probably filter on if its a credit card or not 
		// actually ed wants checking for rent tracking 
		// actually kale uses his debit card smh
		transactions = append(transactions, getReq(account.Links.Transactions, accessToken, client))
	}

	return transactions
}
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env:", err)
		return
	}
	
	certFile := "certs/certificate.pem"
	keyFile := "certs/private_key.pem"
	accessToken := os.Getenv("ACCESS_TOKEN")

	getClient := func(certFile string, keyFile string) *http.Client {
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
	}

	// this should be a singleton but singleton in go looks scary and none of those words are in cse 11
	client := getClient(certFile, keyFile)

	// need to figure out logging at some pt lol

	// --------------------------------

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	transactions := getTransactions(accessToken, client)

	r.Get("/transactions", func(w http.ResponseWriter, r *http.Request) {
		for _, transaction := range transactions {
			w.Write(transaction)
		}
	})

	fmt.Println("http://localhost:3000/transactions") 	
	http.ListenAndServe(":3000", r)
	
}
