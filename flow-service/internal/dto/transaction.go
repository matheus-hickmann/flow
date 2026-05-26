package dto

import "github.com/shopspring/decimal"

// EntryRequest is one debit/credit leg in a posted transaction.
type EntryRequest struct {
	AccountID string          `json:"accountId"`
	Amount    decimal.Decimal `json:"amount"`
	Type      string          `json:"type"` // "DEBIT" or "CREDIT"
}

// EntryResponse mirrors the entry stored in DynamoDB.
type EntryResponse struct {
	ID        string          `json:"id"`
	AccountID string          `json:"accountId"`
	Amount    decimal.Decimal `json:"amount"`
	Type      string          `json:"type"`
}

// PostTransactionRequest is the body of POST /api/v1/ledger/transactions.
// Single-entry requests are expanded server-side: the matching system account
// (Entrada/Saída) receives the counter-leg.
type PostTransactionRequest struct {
	Description   string         `json:"description"`
	ReferenceID   string         `json:"referenceId,omitempty"`
	Category      string         `json:"category,omitempty"`
	BudgetLimitID string         `json:"budgetLimitId,omitempty"`
	Entries       []EntryRequest `json:"entries"`
}

// TransactionResponse is returned after a successful POST.
type TransactionResponse struct {
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Timestamp   string          `json:"timestamp"`
	ReferenceID string          `json:"referenceId,omitempty"`
	Entries     []EntryResponse `json:"entries"`
}

// TransactionListItem is the list shape returned by GET /api/v1/ledger/transactions.
type TransactionListItem struct {
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Timestamp   string          `json:"timestamp"`
	ReferenceID string          `json:"referenceId"`
	Category    string          `json:"category"`
	Entries     []EntryResponse `json:"entries"`
}
