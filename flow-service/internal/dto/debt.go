package dto

import "github.com/shopspring/decimal"

// DebtRequest is the payload for creating a new debt.
type DebtRequest struct {
	Name         string           `json:"name"`
	Amount       *decimal.Decimal `json:"amount"`
	Type         string           `json:"type"` // TO_PAY | TO_RECEIVE
	Counterparty string           `json:"counterparty"`
	DueDate      string           `json:"dueDate,omitempty"` // YYYY-MM-DD, optional
	Notes        string           `json:"notes,omitempty"`
}

// DebtPaymentRequest records a partial or full payment against a debt.
type DebtPaymentRequest struct {
	Amount *decimal.Decimal `json:"amount"`
}

// DebtResponse is returned by list/get endpoints.
type DebtResponse struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Amount       decimal.Decimal `json:"amount"`
	Remaining    decimal.Decimal `json:"remaining"`
	Type         string          `json:"type"`
	Counterparty string          `json:"counterparty"`
	DueDate      string          `json:"dueDate,omitempty"`
	Notes        string          `json:"notes,omitempty"`
	Status       string          `json:"status"` // ACTIVE | SETTLED
	CreatedAt    string          `json:"createdAt"`
}
