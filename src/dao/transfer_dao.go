package dao

import (
	"fmt"
	"sync"

	"github.com/banking/ai-agents-banking/src/models"
)

type TransferDAO struct {
	transfers     map[string]models.Transfer // transferID -> Transfer
	userTransfers map[string][]string        // userID -> []transferID
	mu            sync.RWMutex
}

func NewTransferDAO() *TransferDAO {
	return &TransferDAO{
		transfers:     make(map[string]models.Transfer),
		userTransfers: make(map[string][]string),
	}
}

func (d *TransferDAO) AddTransfer(userID string, transfer models.Transfer) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.transfers[transfer.TransferID] = transfer
	d.userTransfers[userID] = append(d.userTransfers[userID], transfer.TransferID)
	return nil
}

func (d *TransferDAO) GetUserTransfers(userID string) ([]models.Transfer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	transferIDs, exists := d.userTransfers[userID]
	if !exists {
		return []models.Transfer{}, nil
	}

	transfers := make([]models.Transfer, 0, len(transferIDs))
	for _, id := range transferIDs {
		if transfer, exists := d.transfers[id]; exists {
			transfers = append(transfers, transfer)
		}
	}

	return transfers, nil
}

func (d *TransferDAO) GetTransfer(userID, transferID string) (*models.Transfer, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Check if user has this transfer
	transferIDs, exists := d.userTransfers[userID]
	if !exists {
		return nil, fmt.Errorf("no transfers found for user")
	}

	for _, id := range transferIDs {
		if id == transferID {
			if transfer, exists := d.transfers[id]; exists {
				return &transfer, nil
			}
		}
	}

	return nil, fmt.Errorf("transfer not found")
}

func (d *TransferDAO) UpdateTransferStatus(userID, transferID, status string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if user has this transfer
	transferIDs, exists := d.userTransfers[userID]
	if !exists {
		return fmt.Errorf("no transfers found for user")
	}

	for _, id := range transferIDs {
		if id == transferID {
			if transfer, exists := d.transfers[id]; exists {
				transfer.Status = status
				d.transfers[id] = transfer
				return nil
			}
		}
	}

	return fmt.Errorf("transfer not found")
}
