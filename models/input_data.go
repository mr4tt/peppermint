package models

type User struct {
	Username string `json:"username"`
	Password string `json:"pw"`
}

type SalaryInfo struct {
	K401       float64 `json:"contribution_401k"`
	Insurance  float64 `json:"total_insurance_amount"`
	PostTaxSal float64 `json:"monthly_posttax_salary"`
}

type RecurringCost struct {
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	Frequency int64   `json:"frequency"`
	IsSavings bool    `json:"is_savings"`
}

type OneTimeCost struct {
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
	Month  int    `json:"month"`
	Year   int    `json:"year"`
}

type CategoryInfo struct {
	Name  string  `json:"name"`
	Limit float64 `json:"category_limit"`
}
