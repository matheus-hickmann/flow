package dto

import "github.com/shopspring/decimal"

// SummaryResponse is GET /api/v1/dashboard/summary.
type SummaryResponse struct {
	TotalBalance      decimal.Decimal `json:"totalBalance"`
	AccountsCount     int             `json:"accountsCount"`
	InvestmentBalance decimal.Decimal `json:"investmentBalance"`
}

// CategoryAmount is one entry in the monthly category breakdown.
type CategoryAmount struct {
	Category string          `json:"category"`
	Amount   decimal.Decimal `json:"amount"`
}

// MonthlySummaryResponse is GET /api/v1/dashboard/summary/monthly.
type MonthlySummaryResponse struct {
	MonthlyIncome     decimal.Decimal  `json:"monthlyIncome"`
	MonthlyExpense    decimal.Decimal  `json:"monthlyExpense"`
	CategoryBreakdown []CategoryAmount `json:"categoryBreakdown"`
}

// PlannedVsActualItem is one row of GET /api/v1/dashboard/planned-vs-actual.
type PlannedVsActualItem struct {
	Category string          `json:"category"`
	Planned  decimal.Decimal `json:"planned"`
	Actual   decimal.Decimal `json:"actual"`
}

// BalanceTrendPoint is one daily snapshot of the user's total balance.
// Series ordering: oldest → newest, last point being today's end-of-day total.
type BalanceTrendPoint struct {
	Date         string          `json:"date"` // yyyy-MM-dd, UTC
	TotalBalance decimal.Decimal `json:"totalBalance"`
}

// DailyExpensesResponse is GET /api/v1/dashboard/daily-expenses?year=&month=.
// dayTotals is a dense array indexed by day-1 (index 0 = day 1).
type DailyExpensesResponse struct {
	Year             int               `json:"year"`
	Month            int               `json:"month"` // 0-based to match JS Date contract
	DaysInMonth      int               `json:"daysInMonth"`
	DayTotals        []decimal.Decimal `json:"dayTotals"`
	TransactionCount int               `json:"transactionCount"`
}

// BudgetProjectionResponse is GET /api/v1/dashboard/budget-projection?months=N.
// It projects the user's balance month-by-month using their configured budget
// limits and salary. HasBudgets=false means the user has no budgets set up.
type BudgetProjectionResponse struct {
	CurrentBalance decimal.Decimal        `json:"currentBalance"`
	MonthlyBudget  decimal.Decimal        `json:"monthlyBudget"`
	MonthlySalary  decimal.Decimal        `json:"monthlySalary"`
	HasBudgets     bool                   `json:"hasBudgets"`
	HasSalary      bool                   `json:"hasSalary"`
	DaysRemaining  int                    `json:"daysRemaining"`
	Months         []ProjectionMonthResult `json:"months"`
}

// ProjectionMonthResult is one month entry in BudgetProjectionResponse.
// For the current month only the remaining portion of the month is projected.
// Month is 0-based to match the JS Date contract used elsewhere.
type ProjectionMonthResult struct {
	Year             int             `json:"year"`
	Month            int             `json:"month"` // 0-based
	Label            string          `json:"label"` // e.g. "Mai/2026"
	ProjectedIncome  decimal.Decimal `json:"projectedIncome"`
	ProjectedExpense decimal.Decimal `json:"projectedExpense"`
	Delta            decimal.Decimal `json:"delta"`
	RunningBalance   decimal.Decimal `json:"runningBalance"`
	IsCurrent        bool            `json:"isCurrent"`
}

// ProjectionResponse is GET /api/v1/dashboard/projection?year=&month=.
//
// Projection rules (UTC):
//  - Current month: linear extrapolation of current income/expense across the
//    month, optionally adding a configured salary if its day-of-month is still
//    ahead. Basis explains which path was taken.
//  - Past month: returns actual end-of-month figures; daysRemaining=0; basis="past".
//  - Future month: returns currentBalance with zero futures; basis="future".
type ProjectionResponse struct {
	ProjectedBalance decimal.Decimal `json:"projectedBalance"`
	ProjectedIncome  decimal.Decimal `json:"projectedIncome"`
	ProjectedExpense decimal.Decimal `json:"projectedExpense"`
	DaysRemaining    int             `json:"daysRemaining"`
	Basis            string          `json:"basis"` // "linear" | "linear-with-salary" | "past" | "future"
}
