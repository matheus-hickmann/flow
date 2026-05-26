package dto

import "github.com/shopspring/decimal"

// PlanningSubmitRequest is a discriminated payload: only the fields matching
// `Type` are read by the service. Mirrors the Java sealed-ish DTO.
type PlanningSubmitRequest struct {
	Type                string           `json:"type"` // limit | goal | params | salary
	Category            string           `json:"category,omitempty"`
	LimitType           string           `json:"limitType,omitempty"`
	LimitValue          *decimal.Decimal `json:"limitValue,omitempty"`
	Name                string           `json:"name,omitempty"`
	ExpectedReturnRate  *decimal.Decimal `json:"expectedReturnRate,omitempty"`
	MonthlyContribution *decimal.Decimal `json:"monthlyContribution,omitempty"`
	TargetAmount        *decimal.Decimal `json:"targetAmount,omitempty"`
	SelicRate           *decimal.Decimal `json:"selicRate,omitempty"`
	IpcaRate            *decimal.Decimal `json:"ipcaRate,omitempty"`
	Amount              *decimal.Decimal `json:"amount,omitempty"`
	DayOfMonth          *int             `json:"dayOfMonth,omitempty"`
	AccountID           string           `json:"accountId,omitempty"`
}

// BudgetResponse is one row of GET /api/v1/planning/budgets.
type BudgetResponse struct {
	ID         string          `json:"id"`
	Category   string          `json:"category"`
	LimitType  string          `json:"limitType"`
	LimitValue decimal.Decimal `json:"limitValue"`
}

// GoalResponse is one row of GET /api/v1/planning/goals.
type GoalResponse struct {
	ID                  string           `json:"id"`
	Name                string           `json:"name"`
	ExpectedReturnRate  *decimal.Decimal `json:"expectedReturnRate,omitempty"`
	MonthlyContribution *decimal.Decimal `json:"monthlyContribution,omitempty"`
	TargetAmount        *decimal.Decimal `json:"targetAmount,omitempty"`
}

// EconomicParametersResponse is GET /api/v1/planning/economic-parameters.
type EconomicParametersResponse struct {
	SelicRate decimal.Decimal `json:"selicRate"`
	IpcaRate  decimal.Decimal `json:"ipcaRate"`
}

// SalaryResponse is GET /api/v1/planning/salary (204 No Content when absent).
type SalaryResponse struct {
	Amount     decimal.Decimal `json:"amount"`
	DayOfMonth *int            `json:"dayOfMonth,omitempty"`
	AccountID  string          `json:"accountId,omitempty"`
}
