package dao

import (
	"fmt"
	"sync"

	"github.com/banking/ai-agents-banking/src/models"
)

type PayeeDAO struct {
	userPayees map[string][]models.Payee
	mu         sync.RWMutex
}

func NewPayeeDAO() *PayeeDAO {
	return &PayeeDAO{
		userPayees: make(map[string][]models.Payee),
	}
}

func (d *PayeeDAO) GetUserPayees(userID string) ([]models.Payee, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if payees, exists := d.userPayees[userID]; exists {
		return payees, nil
	}

	return []models.Payee{}, nil
}

func (d *PayeeDAO) AddUserPayee(userID string, payee models.Payee) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if payees, exists := d.userPayees[userID]; exists {
		d.userPayees[userID] = append(payees, payee)
	} else {
		d.userPayees[userID] = []models.Payee{payee}
	}

	return nil
}

func (d *PayeeDAO) GetPayeeByAccount(userID, accountNumber string) (*models.Payee, error) {
	payees, err := d.GetUserPayees(userID)
	if err != nil {
		return nil, err
	}

	for _, payee := range payees {
		if payee.AccountNo == accountNumber {
			return &payee, nil
		}
	}

	return nil, nil
}

func (d *PayeeDAO) GetUserPayee(userID, payeeID string) (*models.Payee, error) {
	payees, err := d.GetUserPayees(userID)
	if err != nil {
		return nil, err
	}

	for _, payee := range payees {
		if payee.ID == payeeID {
			return &payee, nil
		}
	}

	return nil, fmt.Errorf("payee not found")
}

func (d *PayeeDAO) UpdateUserPayee(userID string, payee models.Payee) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	payees, exists := d.userPayees[userID]
	if !exists {
		return fmt.Errorf("no payees found for user")
	}

	for i, p := range payees {
		if p.ID == payee.ID {
			payees[i] = payee
			d.userPayees[userID] = payees
			return nil
		}
	}

	return fmt.Errorf("payee not found")
}

func (d *PayeeDAO) DeleteUserPayee(userID, payeeID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	payees, exists := d.userPayees[userID]
	if !exists {
		return fmt.Errorf("no payees found for user")
	}

	for i, payee := range payees {
		if payee.ID == payeeID {
			// Remove the payee by appending the slices before and after it
			d.userPayees[userID] = append(payees[:i], payees[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("payee not found")
}
