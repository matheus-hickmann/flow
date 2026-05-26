package dto

import "github.com/shopspring/decimal"

// ParsedRow is one transaction row extracted from a CSV statement.
type ParsedRow struct {
	Date          string          `json:"date"`          // YYYY-MM-DD
	Description   string          `json:"description"`
	Amount        decimal.Decimal `json:"amount"`        // always positive
	Type          string          `json:"type"`          // "DEBIT" (expense) or "CREDIT" (income)
	Category      string          `json:"category"`      // filled when a rule matches
	NeedsCategory bool            `json:"needsCategory"` // true when no rule found
	MerchantKey   string          `json:"merchantKey"`   // normalized key used for rule lookup
}

// MerchantRule maps a normalized merchant key to a user-chosen category.
type MerchantRule struct {
	MerchantKey string `json:"merchantKey"`
	DisplayName string `json:"displayName"` // human-readable (first matched description)
	Category    string `json:"category"`
}

// ImportPreviewResponse is returned by POST /api/v1/imports/parse.
type ImportPreviewResponse struct {
	Rows       []ParsedRow    `json:"rows"`
	KnownRules []MerchantRule `json:"knownRules"`
}

// ImportCommitRequest is the body of POST /api/v1/imports/commit.
type ImportCommitRequest struct {
	AccountID     string         `json:"accountId"`
	Rows          []ParsedRow    `json:"rows"`          // all rows with categories resolved
	MerchantRules []MerchantRule `json:"merchantRules"` // new/updated rules to persist
}

// ImportCommitResponse summarises the result of a commit.
type ImportCommitResponse struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}