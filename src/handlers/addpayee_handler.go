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

type PayeeHandler struct {
	payeeDAO *dao.PayeeDAO
}

func NewPayeeHandler(payeeDAO *dao.PayeeDAO) *PayeeHandler {
	return &PayeeHandler{
		payeeDAO: payeeDAO,
	}
}

func (h *PayeeHandler) ListPayees(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	payees, err := h.payeeDAO.GetUserPayees(userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch payees"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"payees":  payees,
		"user_id": userID,
		"total":   len(payees),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PayeeHandler) CreatePayee(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Name      string `json:"name"`
		AccountNo string `json:"account_no"`
		IFSCCode  string `json:"ifsc_code"`
		BankName  string `json:"bank_name,omitempty"`
		UPIId     string `json:"upi_id,omitempty"`
		NickName  string `json:"nick_name,omitempty"`
		PayeeType string `json:"payee_type,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	payee := models.Payee{
		ID:         time.Now().Format("PAYEE_20060102150405"),
		Name:       req.Name,
		AccountNo:  req.AccountNo,
		BankName:   req.BankName,
		IFSCCode:   req.IFSCCode,
		PayeeType:  req.PayeeType,
		UPIId:      req.UPIId,
		AddedDate:  time.Now(),
		IsActive:   true,
		IsVerified: false,
		NickName:   req.NickName,
	}

	err := h.payeeDAO.AddUserPayee(userID, payee)
	if err != nil {
		http.Error(w, `{"error": "Failed to add payee"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payee)
}

func (h *PayeeHandler) GetPayee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payeeID := vars["payeeId"]

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	payee, err := h.payeeDAO.GetUserPayee(userID, payeeID)
	if err != nil {
		http.Error(w, `{"error": "Payee not found"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"payee":   payee,
		"user_id": userID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PayeeHandler) UpdatePayee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payeeID := vars["payeeId"]

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		Name     string `json:"name,omitempty"`
		NickName string `json:"nick_name,omitempty"`
		IsActive *bool  `json:"is_active,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	payee, err := h.payeeDAO.GetUserPayee(userID, payeeID)
	if err != nil {
		http.Error(w, `{"error": "Payee not found"}`, http.StatusNotFound)
		return
	}

	if req.Name != "" {
		payee.Name = req.Name
	}
	if req.NickName != "" {
		payee.NickName = req.NickName
	}
	if req.IsActive != nil {
		payee.IsActive = *req.IsActive
	}

	err = h.payeeDAO.UpdateUserPayee(userID, *payee)
	if err != nil {
		http.Error(w, `{"error": "Failed to update payee"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"payee":   payee,
		"user_id": userID,
		"message": "Payee updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PayeeHandler) DeletePayee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	payeeID := vars["payeeId"]

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	err := h.payeeDAO.DeleteUserPayee(userID, payeeID)
	if err != nil {
		http.Error(w, `{"error": "Failed to delete payee"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"payee_id": payeeID,
		"user_id":  userID,
		"message":  "Payee deleted successfully",
		"deleted":  true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
