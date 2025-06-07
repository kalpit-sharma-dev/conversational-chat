package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/banking/ai-agents-banking/src/dao"
	"github.com/banking/ai-agents-banking/src/models"
)

type TransferHandler struct {
	transferDAO *dao.TransferDAO
	accountDAO  *dao.AccountDAO
}

func NewTransferHandler(transferDAO *dao.TransferDAO, accountDAO *dao.AccountDAO) *TransferHandler {
	return &TransferHandler{
		transferDAO: transferDAO,
		accountDAO:  accountDAO,
	}
}

func (h *TransferHandler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	transfers, err := h.transferDAO.GetUserTransfers(userID)
	if err != nil {
		http.Error(w, "Failed to list transfers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfers)
}

func (h *TransferHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var transfer models.Transfer
	if err := json.NewDecoder(r.Body).Decode(&transfer); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate source account belongs to user
	sourceAccount, err := h.accountDAO.GetAccountByID(userID, transfer.FromAccountID)
	if err != nil {
		http.Error(w, "Invalid source account", http.StatusBadRequest)
		return
	}

	if sourceAccount.Balance < transfer.Amount {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Create transfer
	if err := h.transferDAO.AddTransfer(userID, transfer); err != nil {
		http.Error(w, "Failed to create transfer", http.StatusInternalServerError)
		return
	}

	// Update account balances
	if err := h.accountDAO.UpdateAccountBalance(userID, transfer.FromAccountID, sourceAccount.Balance-transfer.Amount); err != nil {
		http.Error(w, "Failed to update source account", http.StatusInternalServerError)
		return
	}

	destAccount, err := h.accountDAO.GetAccountByID(userID, transfer.ToAccountID)
	if err != nil {
		http.Error(w, "Invalid destination account", http.StatusBadRequest)
		return
	}

	if err := h.accountDAO.UpdateAccountBalance(userID, transfer.ToAccountID, destAccount.Balance+transfer.Amount); err != nil {
		http.Error(w, "Failed to update destination account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfer)
}

func (h *TransferHandler) GetTransfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	transferID := r.URL.Query().Get("id")
	if transferID == "" {
		http.Error(w, "Transfer ID required", http.StatusBadRequest)
		return
	}

	transfer, err := h.transferDAO.GetTransfer(userID, transferID)
	if err != nil {
		http.Error(w, "Transfer not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transfer)
}
