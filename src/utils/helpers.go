package utils

import "strings"

func CalculateTransferFees(method string, amount float64) float64 {
	switch strings.ToUpper(method) {
	case "UPI":
		return 0.0 // UPI is free
	case "IMPS":
		if amount <= 10000 {
			return 5.0
		} else if amount <= 100000 {
			return 15.0
		}
		return 25.0
	case "NEFT":
		if amount <= 10000 {
			return 2.5
		} else if amount <= 100000 {
			return 5.0
		}
		return 15.0
	case "RTGS":
		if amount <= 200000 {
			return 25.0
		}
		return 50.0
	default:
		return 10.0
	}
}

func GetBankNameFromIFSC(ifsc string) string {
	bankCodes := map[string]string{
		"SBIN": "State Bank of India",
		"HDFC": "HDFC Bank",
		"ICIC": "ICICI Bank",
		"AXIS": "Axis Bank",
		"PUNB": "Punjab National Bank",
		"UBIN": "Union Bank of India",
		"CNRB": "Canara Bank",
		"BARB": "Bank of Baroda",
	}

	if len(ifsc) >= 4 {
		if bankName, exists := bankCodes[ifsc[:4]]; exists {
			return bankName
		}
	}
	return "Unknown Bank"
}
