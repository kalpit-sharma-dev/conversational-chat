package dao

import (
	"fmt"
	"math"
	"sync"

	"github.com/banking/ai-agents-banking/src/models"
)

type LoanDAO struct {
	loanProducts     []models.LoanProduct
	applications     map[string]models.LoanApplication
	userApplications map[string][]string // userID -> []applicationID
	mu               sync.RWMutex
}

func NewLoanDAO() *LoanDAO {
	dao := &LoanDAO{
		applications:     make(map[string]models.LoanApplication),
		userApplications: make(map[string][]string),
	}
	dao.initializeLoanProducts()
	return dao
}

func (d *LoanDAO) GetLoanProducts() ([]models.LoanProduct, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.loanProducts, nil
}

func (d *LoanDAO) GetLoanProduct(loanType string) (*models.LoanProduct, error) {
	products, err := d.GetLoanProducts()
	if err != nil {
		return nil, err
	}

	for _, product := range products {
		if product.LoanType == loanType {
			return &product, nil
		}
	}

	return nil, nil
}

func (d *LoanDAO) GetUserLoanApplications(userID string) ([]models.LoanApplication, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	applicationIDs, exists := d.userApplications[userID]
	if !exists {
		return []models.LoanApplication{}, nil
	}

	applications := make([]models.LoanApplication, 0, len(applicationIDs))
	for _, id := range applicationIDs {
		if app, exists := d.applications[id]; exists {
			applications = append(applications, app)
		}
	}

	return applications, nil
}

func (d *LoanDAO) CreateLoanApplication(userID string, application models.LoanApplication) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.applications[application.ApplicationID] = application
	d.userApplications[userID] = append(d.userApplications[userID], application.ApplicationID)
	return nil
}

func (d *LoanDAO) GetLoanApplication(userID, applicationID string) (*models.LoanApplication, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Check if user has this application
	applicationIDs, exists := d.userApplications[userID]
	if !exists {
		return nil, fmt.Errorf("no applications found for user")
	}

	for _, id := range applicationIDs {
		if id == applicationID {
			if app, exists := d.applications[id]; exists {
				return &app, nil
			}
		}
	}

	return nil, fmt.Errorf("application not found")
}

func (d *LoanDAO) CheckLoanEligibility(userID, loanType string, amount, income float64) (*models.LoanEligibility, error) {
	product, err := d.GetLoanProduct(loanType)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, fmt.Errorf("invalid loan type")
	}

	eligibility := &models.LoanEligibility{
		Eligible: true,
		Reason:   "Eligible for loan",
	}

	// Check amount limits
	if amount < product.MinAmount {
		eligibility.Eligible = false
		eligibility.Reason = fmt.Sprintf("Amount below minimum limit of ₹%.2f", product.MinAmount)
		return eligibility, nil
	}

	if amount > product.MaxAmount {
		eligibility.Eligible = false
		eligibility.Reason = fmt.Sprintf("Amount above maximum limit of ₹%.2f", product.MaxAmount)
		return eligibility, nil
	}

	// Check income criteria
	emi, err := d.CalculateEMI(amount, product.InterestRate, 60) // Assuming 5 years tenure for eligibility
	if err != nil {
		return nil, err
	}
	if emi > income*0.4 { // EMI should not exceed 40% of income
		eligibility.Eligible = false
		eligibility.Reason = "EMI exceeds 40% of income"
		return eligibility, nil
	}

	return eligibility, nil
}

func (d *LoanDAO) CalculateEMI(amount, rate float64, tenure int) (float64, error) {
	if amount <= 0 || rate <= 0 || tenure <= 0 {
		return 0, fmt.Errorf("invalid parameters")
	}

	// Convert annual rate to monthly rate
	monthlyRate := rate / 12 / 100

	// Calculate EMI using the formula: EMI = P * r * (1 + r)^n / ((1 + r)^n - 1)
	// where P = principal, r = monthly rate, n = number of months
	power := math.Pow(1+monthlyRate, float64(tenure))
	emi := amount * monthlyRate * power / (power - 1)

	return math.Round(emi*100) / 100, nil
}

func (d *LoanDAO) initializeLoanProducts() {
	d.loanProducts = []models.LoanProduct{
		{
			LoanType:            "personal",
			Name:                "Personal Loan",
			Description:         "Instant personal loan for any purpose",
			MinAmount:           50000,
			MaxAmount:           2500000,
			InterestRate:        12.0,
			MaxTenure:           60,
			ProcessingFee:       2.0,
			EligibilityCriteria: []string{"Age 21-60", "Minimum income ₹25,000", "Good credit score"},
		},
		{
			LoanType:            "home",
			Name:                "Home Loan",
			Description:         "Loan for purchasing or constructing home",
			MinAmount:           500000,
			MaxAmount:           50000000,
			InterestRate:        8.5,
			MaxTenure:           300,
			ProcessingFee:       0.5,
			EligibilityCriteria: []string{"Age 21-65", "Property papers", "Income proof"},
		},
		{
			LoanType:            "car",
			Name:                "Car Loan",
			Description:         "Loan for purchasing new or used cars",
			MinAmount:           100000,
			MaxAmount:           5000000,
			InterestRate:        9.5,
			MaxTenure:           84,
			ProcessingFee:       1.0,
			EligibilityCriteria: []string{"Age 21-65", "Valid driving license", "Income proof"},
		},
	}
}
