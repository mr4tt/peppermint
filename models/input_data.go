package models

type User struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"pw" validate:"required"`
}

type SalaryInfo struct {
	K401       float64 `json:"contribution_401k" validate:"required"`
	Insurance  float64 `json:"total_insurance_amount" validate:"required"`
	PostTaxSal float64 `json:"monthly_posttax_salary" validate:"required"`
}

type RecurringCost struct {
	Name      string  `json:"name" validate:"required"`
	Amount    float64 `json:"amount" validate:"required"`
	Frequency int64   `json:"frequency" validate:"required"`
	IsSavings bool    `json:"is_savings" validate:"required"`
}

type OneTimeCost struct {
	Name   string `json:"name" validate:"required"`
	Amount int64  `json:"amount" validate:"required"`
	Month  int    `json:"month" validate:"required"`
	Year   int    `json:"year" validate:"required"`
}

type CategoryInfo struct {
	Name  string  `json:"name" validate:"required"`
	Limit float64 `json:"category_limit" validate:"required"`
}
