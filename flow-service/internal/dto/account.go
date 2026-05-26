package dto

import "github.com/shopspring/decimal"

// AccountResponse mirrors the Java AccountResponse record.
// All optional fields omit when zero/null to keep payloads small.
type AccountResponse struct {
	ID          string           `json:"id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Balance     decimal.Decimal  `json:"balance"`
	Color       string           `json:"color"`
	IsSystem    bool             `json:"isSystem"`
	Investment  bool             `json:"investment"`
	Shared      bool             `json:"shared"`
	AnnualRate  *decimal.Decimal `json:"annualRate,omitempty"`
	Brand       string           `json:"brand,omitempty"`
	CreditLimit *decimal.Decimal `json:"limit,omitempty"`
	ClosingDay  *int             `json:"closingDay,omitempty"`
	DueDay      *int             `json:"dueDay,omitempty"`
}

// CreateAccountRequest is the body of POST /api/v1/ledger/accounts.
type CreateAccountRequest struct {
	Name           string           `json:"name"`
	InitialBalance decimal.Decimal  `json:"initialBalance"`
	Color          string           `json:"color"`
	System         bool             `json:"system"`
	Investment     bool             `json:"investment"`
	AnnualRate     *decimal.Decimal `json:"annualRate,omitempty"`
	Brand          string           `json:"brand,omitempty"`
	Limit          *decimal.Decimal `json:"limit,omitempty"`
	ClosingDay     *int             `json:"closingDay,omitempty"`
	DueDay         *int             `json:"dueDay,omitempty"`
}

// ColorOrDefault returns the requested color or a fallback blue.
func (r CreateAccountRequest) ColorOrDefault() string {
	if r.Color == "" {
		return "#3b82f6"
	}
	return r.Color
}

// UpdateAccountRequest is the body of PATCH /api/v1/ledger/accounts/{id}.
// Nil/zero fields mean "do not change".
type UpdateAccountRequest struct {
	Name       string           `json:"name,omitempty"`
	Color      string           `json:"color,omitempty"`
	Investment *bool            `json:"investment,omitempty"`
	Shared     *bool            `json:"shared,omitempty"`
	AnnualRate *decimal.Decimal `json:"annualRate,omitempty"`
	Brand      string           `json:"brand,omitempty"`
	Limit      *decimal.Decimal `json:"limit,omitempty"`
	ClosingDay *int             `json:"closingDay,omitempty"`
	DueDay     *int             `json:"dueDay,omitempty"`
}

// AdjustBalanceRequest is the body of PATCH /api/v1/ledger/accounts/{id}/balance.
type AdjustBalanceRequest struct {
	NewBalance decimal.Decimal `json:"newBalance"`
}

// BalanceItem mirrors the Java BalanceItem record (BALANCE#LATEST snapshot).
type BalanceItem struct {
	CurrentBalance decimal.Decimal `json:"currentBalance"`
	LastUpdate     string          `json:"lastUpdate"`
}
