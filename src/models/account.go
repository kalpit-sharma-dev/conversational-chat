package models

import "time"

// Account represents a bank account
type Account struct {
	AccountID      string    `json:"account_id"`
	UserID         string    `json:"user_id"`
	AccountNumber  string    `json:"account_number"`
	AccountType    string    `json:"account_type"` // Savings, Current, Fixed Deposit, etc.
	Balance        float64   `json:"balance"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"` // Active, Inactive, Frozen, etc.
	OpeningDate    time.Time `json:"opening_date"`
	LastUpdated    time.Time `json:"last_updated"`
	BranchCode     string    `json:"branch_code,omitempty"`
	IFSCCode       string    `json:"ifsc_code,omitempty"`
	InterestRate   float64   `json:"interest_rate,omitempty"`
	MaturityDate   time.Time `json:"maturity_date,omitempty"`
	MinimumBalance float64   `json:"minimum_balance,omitempty"`
}
