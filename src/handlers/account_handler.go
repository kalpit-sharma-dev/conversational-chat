package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/middleware"
)

type AccountHandler struct {
	accountDAO *dao.AccountDAO
}

func NewAccountHandler(accountDAO *dao.AccountDAO) *AccountHandler {
	return &AccountHandler{
		accountDAO: accountDAO,
	}
}

func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	accounts, err := h.accountDAO.GetUserAccounts(userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to retrieve accounts"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["accountId"]

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	account, err := h.accountDAO.GetAccountByID(userID, accountID)
	if err != nil {
		http.Error(w, `{"error": "Failed to retrieve account"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

func (h *AccountHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["accountId"]

	userID, ok := middleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error": "User not found in context"}`, http.StatusUnauthorized)
		return
	}

	account, err := h.accountDAO.GetAccountByID(userID, accountID)
	if err != nil {
		http.Error(w, `{"error": "Failed to retrieve account balance"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"account_id": accountID,
		"balance":    account.Balance,
		"currency":   account.Currency,
		"status":     account.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
