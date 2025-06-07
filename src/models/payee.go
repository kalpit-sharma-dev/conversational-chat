package models

import "time"

// Payee represents a beneficiary/payee in the banking system
type Payee struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	AccountNo  string    `json:"account_no"`
	BankName   string    `json:"bank_name,omitempty"`
	IFSCCode   string    `json:"ifsc_code,omitempty"`
	PayeeType  string    `json:"payee_type"`
	UPIId      string    `json:"upi_id,omitempty"`
	AddedDate  time.Time `json:"added_date"`
	IsActive   bool      `json:"is_active"`
	IsVerified bool      `json:"is_verified"`
	NickName   string    `json:"nick_name,omitempty"`
}
