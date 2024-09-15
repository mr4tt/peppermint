package models

// response: https://teller.io/docs/api/identity#get-identity
// links: https://teller.io/docs/api/accounts#properties
type CapitalOneResp struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Status  string `json:"status"`
	Name    string `json:"name"`
	Links   struct {
		Transactions string `json:"transactions"`
		Self         string `json:"self"`
		Balances     string `json:"balances"`
	} `json:"links"`
	Institution struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"institution"`
	ID           string `json:"id"`
	EnrollmentID string `json:"enrollment_id"`
}

// https://teller.io/docs/api/account/transactions
type Transaction struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Links  struct {
		Self    string `json:"self"`
		Account string `json:"account"`
	} `json:"links"`
	ID      string `json:"id"`
	Details struct {
		ProcessingStatus string `json:"processing_status"`
		Counterparty     struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"counterparty"`
		Category string `json:"category"`
	} `json:"details"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Amount      string `json:"amount"`
	AccountID   string `json:"account_id"`
}
