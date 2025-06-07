package models

import "time"

// LoanProduct represents a loan product offered by the bank
type LoanProduct struct {
	LoanType            string   `json:"loan_type"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	MinAmount           float64  `json:"min_amount"`
	MaxAmount           float64  `json:"max_amount"`
	InterestRate        float64  `json:"interest_rate"`
	MaxTenure           int      `json:"max_tenure"`
	ProcessingFee       float64  `json:"processing_fee"`
	EligibilityCriteria []string `json:"eligibility_criteria"`
}

// LoanApplication represents a loan application submitted by a user
type LoanApplication struct {
	ApplicationID string    `json:"application_id"`
	LoanType      string    `json:"loan_type"`
	Amount        float64   `json:"amount"`
	Tenure        int       `json:"tenure"`
	Status        string    `json:"status"`
	AppliedDate   time.Time `json:"applied_date"`
	EMI           float64   `json:"emi,omitempty"`
	Purpose       string    `json:"purpose"`
	Documents     []string  `json:"documents,omitempty"`
}

// LoanEligibility represents the eligibility check result for a loan
type LoanEligibility struct {
	Eligible bool   `json:"eligible"`
	Reason   string `json:"reason"`
}
