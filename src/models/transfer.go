package models

import "time"

// TransferRequest represents a fund transfer request
type TransferRequest struct {
	TransferID  string    `json:"transfer_id"`
	Amount      float64   `json:"amount"`
	Method      string    `json:"method"` // UPI, IMPS, NEFT
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Reference   string    `json:"reference"`
	Fees        float64   `json:"fees"`
	Description string    `json:"description,omitempty"`
}

// Transfer represents a completed fund transfer
type Transfer struct {
	TransferID    string    `json:"transfer_id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        float64   `json:"amount"`
	Method        string    `json:"method"` // UPI, IMPS, NEFT
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
	Reference     string    `json:"reference"`
	Fees          float64   `json:"fees"`
	Description   string    `json:"description,omitempty"`
}
