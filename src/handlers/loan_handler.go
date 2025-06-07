package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/middleware"
	"github.com/banking/ai-agents-banking/src/models"
)

type LoanHandler struct {
	loanDAO *dao.LoanDAO
}

func NewLoanHandler(loanDAO *dao.LoanDAO) *LoanHandler {
	return &LoanHandler{
		loanDAO: loanDAO,
	}
}

func (h *LoanHandler) ListLoanProducts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	products, err := h.loanDAO.GetLoanProducts()
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch loan products"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"products": products,
		"user_id":  userID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LoanHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	applications, err := h.loanDAO.GetUserLoanApplications(userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch loan applications"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"applications": applications,
		"user_id":      userID,
		"total":        len(applications),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LoanHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		LoanType string  `json:"loan_type"`
		Amount   float64 `json:"amount"`
		Tenure   int     `json:"tenure"`
		Purpose  string  `json:"purpose"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	application := models.LoanApplication{
		ApplicationID: time.Now().Format("LOAN_20060102150405"),
		LoanType:      req.LoanType,
		Amount:        req.Amount,
		Tenure:        req.Tenure,
		Status:        "PENDING",
		AppliedDate:   time.Now(),
		Purpose:       req.Purpose,
	}

	err := h.loanDAO.CreateLoanApplication(userID, application)
	if err != nil {
		http.Error(w, `{"error": "Failed to create loan application"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(application)
}

func (h *LoanHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationId"]

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	application, err := h.loanDAO.GetLoanApplication(userID, applicationID)
	if err != nil {
		http.Error(w, `{"error": "Loan application not found"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"application": application,
		"user_id":     userID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LoanHandler) CheckEligibility(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		LoanType string  `json:"loan_type,omitempty"`
		Amount   float64 `json:"amount,omitempty"`
		Income   float64 `json:"income,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	eligibility, err := h.loanDAO.CheckLoanEligibility(userID, req.LoanType, req.Amount, req.Income)
	if err != nil {
		http.Error(w, `{"error": "Failed to check eligibility"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"eligible": eligibility.Eligible,
		"reason":   eligibility.Reason,
		"user_id":  userID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LoanHandler) CalculateEMI(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
		Rate   float64 `json:"rate,omitempty"`
		Tenure int     `json:"tenure,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Rate == 0 {
		req.Rate = 12.0
	}
	if req.Tenure == 0 {
		req.Tenure = 24
	}

	emi, err := h.loanDAO.CalculateEMI(req.Amount, req.Rate, req.Tenure)
	if err != nil {
		http.Error(w, `{"error": "Failed to calculate EMI"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"amount":  req.Amount,
		"rate":    req.Rate,
		"tenure":  req.Tenure,
		"emi":     emi,
		"user_id": userID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
