package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr4tt/peppermint/models"
)

var (
	DatabaseErrorJson = map[string]string{"error": "Database error"}
	Validator         = validator.New(validator.WithRequiredStructEnabled())
)

// load in Teller certs
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

// make a GET request to Teller, given auth, url to request, and a client (for certs)
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

type Handler struct {
	DBPool *pgxpool.Pool
	Token string
}

// Produces a map that can be used to return a JSON error response for invalid input.
func ProduceInputErrorMap(err error) map[string]string {
	// TODO: might be worth not including the details here, but still good for
	// debugging purposes.
	return map[string]string{
		"error":   "Either the body structure is invalid or a field is missing.",
		"details": err.Error(),
	}
}

// Reads the body of the request, and then unmarshals it into the specified
// type as well as checking if all required fields are present.
//
// Returns the unmarshalled object, or an error if any occurred. The error
// should be checked first before using the returned object.
func ReadBodyAndUnmarshal[T any](r *http.Request) (T, error) {
	var result T
	body, readErr := io.ReadAll(r.Body)
	fmt.Println("Raw body received: ", string(body))
	if readErr != nil {
		fmt.Fprintln(os.Stderr, "Failed to read request body: ", readErr)
		return result, fmt.Errorf("failed to read request body")
	}

	if unmarshalErr := json.Unmarshal(body, &result); unmarshalErr != nil {
		fmt.Fprintln(os.Stderr, "Failed to unmarshal request body: ", unmarshalErr)
		return result, fmt.Errorf("failed to unmarshal the request body. Is the JSON syntatically valid?")
	}

	if validatorErr := Validator.Struct(result); validatorErr != nil {
		fmt.Fprintf(os.Stderr, "Unable to validate the request body, '%s'; error: %v\n", body, validatorErr)
		return result, validatorErr
	}

	fmt.Println("Successfully unmarshalled body: ", result)
	return result, nil
}

// Checks if the username in question is already taken.
//
// Returns true if the username is taken, false otherwise.
// Note that an error will be returned if the query fails;
// this should be checked first before using the returned value.
func (b Handler) IsUsernameUsed(username string) (bool, error) {
	var exists bool
	err := b.DBPool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (b Handler) CheckIfUsernameExists(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	exists, err := b.IsUsernameUsed(username)
	if err != nil {
		fmt.Printf("CheckIfUsernameExists query failed: %v\n", err)
		render.JSON(w, r, DatabaseErrorJson)
		render.Status(r, http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]bool{"exists": exists})
}

func (b Handler) SaveUser(w http.ResponseWriter, r *http.Request) {
	userInfo, check_err := ReadBodyAndUnmarshal[models.User](r)
	if check_err != nil {
		render.JSON(w, r, ProduceInputErrorMap(check_err))
		render.Status(r, http.StatusBadRequest)
		return
	}

	// Check if the username already exists.
	// Note: we could technically just not have this and assume that
	// the frontend will handle it.
	exists, check_err := b.IsUsernameUsed(userInfo.Username)
	if check_err != nil {
		fmt.Printf("CheckIfUsernameExists query failed: %v\n", check_err)
		render.JSON(w, r, DatabaseErrorJson)
		render.Status(r, http.StatusInternalServerError)
		return
	}

	if exists {
		render.JSON(w, r, map[string]string{"error": "Username already taken"})
		render.Status(r, http.StatusConflict)
		return
	}

	userQuery := `INSERT INTO users (username, pw_hash) VALUES (@name, @hash)`
	args := pgx.NamedArgs{
		"name": userInfo.Username,
		"hash": userInfo.Password,
	}

	_, err := b.DBPool.Exec(context.Background(), userQuery, args)

	if err != nil {
		fmt.Printf("SaveUser insert query failed: %v\n", err)
		render.JSON(w, r, DatabaseErrorJson)
		render.Status(r, http.StatusInternalServerError)
		return
	}
}

func (b Handler) SaveSalaryInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userInfo, check_err := ReadBodyAndUnmarshal[models.SalaryInfo](r)
	if check_err != nil {
		render.JSON(w, r, ProduceInputErrorMap(check_err))
		render.Status(r, http.StatusBadRequest)
		return
	}

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
		render.JSON(w, r, DatabaseErrorJson)
		render.Status(r, http.StatusInternalServerError)
		return
	}
}

func (b Handler) SaveRecurringCostInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	recurringCosts, check_err := ReadBodyAndUnmarshal[[]models.RecurringCost](r)
	if check_err != nil {
		render.JSON(w, r, ProduceInputErrorMap(check_err))
		render.Status(r, http.StatusBadRequest)
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
			render.JSON(w, r, DatabaseErrorJson)
			render.Status(r, http.StatusInternalServerError)
			return
		}
	}
}

func (b Handler) SaveOneTimeCost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	userCosts, check_err := ReadBodyAndUnmarshal[[]models.OneTimeCost](r)
	if check_err != nil {
		render.JSON(w, r, ProduceInputErrorMap(check_err))
		render.Status(r, http.StatusBadRequest)
		return
	}

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
			render.JSON(w, r, DatabaseErrorJson)
			render.Status(r, http.StatusInternalServerError)
			return
		}
	}
}

func (b Handler) SaveCategories(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	userCategories, check_err := ReadBodyAndUnmarshal[[]models.CategoryInfo](r)
	if check_err != nil {
		render.JSON(w, r, ProduceInputErrorMap(check_err))
		render.Status(r, http.StatusBadRequest)
		return
	}

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
			render.JSON(w, r, DatabaseErrorJson)
			render.Status(r, http.StatusInternalServerError)
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

func (b Handler) GetNewTransactionsFromTeller(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	accessToken := os.Getenv("ACCESS_TOKEN")

	noTransactions := false

	// get the latest transaction saved
	lastPostedDate, err := func(id string) (string, error) {
		getTransactionsQuery := `
		SELECT posted_date FROM transactions 
		WHERE user_id=@id 
		ORDER BY posted_date DESC 
		LIMIT 1`

		args := pgx.NamedArgs{
			"id":  id,
		}
	
		var posted_date time.Time
	
		err := b.DBPool.QueryRow(context.Background(), getTransactionsQuery, args).Scan(&posted_date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
			return "", err
		}
		isoDateString := posted_date.Format("2006-01-02")
		return isoDateString, nil
	}(id)
	
	if errors.Is(err, pgx.ErrNoRows) {
		noTransactions = true
	}

	fmt.Println("noTransaction value: ", noTransactions)
	fmt.Println("lastPostedDate value: ", lastPostedDate)

	var transactions []models.Transaction
	
	if !noTransactions {
		accInfo := getAccounts(accessToken)
		for _, account := range accInfo {
			if account.Subtype != "checking" && account.Subtype != "credit_card" {
				continue
			}

			url := fmt.Sprintf("https://api.teller.io/accounts/%s/transactions?start_date=%s", account.ID, lastPostedDate)
			fmt.Println(url)

			rawTransactions := getReq(url, accessToken)

			// fmt.Println("rawTransactions: ", string(rawTransactions))

			var parsedTransactions []models.Transaction
			err = json.Unmarshal((rawTransactions), &parsedTransactions)
			if err != nil {
				fmt.Println("Error unmarshalling transactions:", err)
				return
			}

			for _, transaction := range parsedTransactions {
				// We only want to process posted transactions
				if transaction.Status != "posted" {
					continue
				}

				transactions = append(transactions, transaction)
			}
		}
	} else {
		transactions = getTransactions(accessToken)
	}
	
	fmt.Println("transactions: ", transactions)


	// args := pgx.NamedArgs{
	// 	"uid":  id,
	// 	"name": category.Name,
	// 	"lim":  category.Limit,
	// }

	// insertQuery := `
	// INSERT INTO my_table (data)
	// VALUES ('[
    // {"name": "Alice", "age": 30},
    // {"name": "Bob", "age": 25},
    // {"name": "Charlie", "age": 35}
	// ]');` and on conflict, do nothing (because the dates mean we'll have txn dups)



	// call teller to get everything after that id if there are previous transactions found
	// unmarshall, save to db
}
func (b Handler) GetTransactionsFromDB(w http.ResponseWriter, r *http.Request) {}

func (b Handler) GetCategories(w http.ResponseWriter, r *http.Request) {}

func (b Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {}
func (b Handler) DeleteCategory(w http.ResponseWriter, r *http.Request)    {}
