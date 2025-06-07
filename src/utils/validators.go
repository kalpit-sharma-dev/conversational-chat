package utils

import (
	"regexp"
	"strings"
)

func IsValidIFSC(ifsc string) bool {
	// IFSC format: 4 letters + 7 characters (letters/numbers)
	matched, _ := regexp.MatchString(`^[A-Z]{4}[0-9A-Z]{7}$`, strings.ToUpper(ifsc))
	return matched
}

func IsValidAccountNumber(accNo string) bool {
	// Account number should be 9-18 digits
	matched, _ := regexp.MatchString(`^[0-9]{9,18}$`, accNo)
	return matched
}
