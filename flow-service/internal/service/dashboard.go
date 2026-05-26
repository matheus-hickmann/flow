package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
)

// DashboardService aggregates accounts + transactions + budgets into the views
// the frontend uses on the dashboard.
type DashboardService struct {
	accounts     *AccountService
	transactions *TransactionService
	budgets      *PlanningService
}

// NewDashboardService wires the service.
func NewDashboardService(accounts *AccountService, transactions *TransactionService, budgets *PlanningService) *DashboardService {
	return &DashboardService{accounts: accounts, transactions: transactions, budgets: budgets}
}

// GetSummary returns the total balance and account count (system accounts excluded).
func (s *DashboardService) GetSummary(ctx context.Context, userID string) (dto.SummaryResponse, error) {
	all, err := s.accounts.ListFiltered(ctx, userID, true)
	if err != nil {
		return dto.SummaryResponse{}, err
	}
	var total, investment decimal.Decimal
	count := 0
	for _, a := range all {
		if a.IsSystem {
			continue
		}
		count++
		total = total.Add(a.Balance)
		if a.Investment {
			investment = investment.Add(a.Balance)
		}
	}
	return dto.SummaryResponse{
		TotalBalance:      total,
		AccountsCount:     count,
		InvestmentBalance: investment,
	}, nil
}

// GetMonthlySummary returns income, expense and per-category expense breakdown
// for a given month (month is 0-based to match the JS Date contract).
func (s *DashboardService) GetMonthlySummary(ctx context.Context, userID string, year, month int) (dto.MonthlySummaryResponse, error) {
	accounts, err := s.accounts.ListFiltered(ctx, userID, true)
	if err != nil {
		return dto.MonthlySummaryResponse{}, err
	}
	incomeID, expenseID := systemAccountIDs(accounts)

	start := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	txns, err := s.transactions.List(ctx, userID, 500, "")
	if err != nil {
		return dto.MonthlySummaryResponse{}, err
	}

	var income, expense decimal.Decimal
	categoryTotals := map[string]decimal.Decimal{}
	categoryOrder := []string{}

	for _, tx := range txns {
		ts, err := time.Parse(time.RFC3339Nano, tx.Timestamp)
		if err != nil {
			continue
		}
		if ts.Before(start) || !ts.Before(end) {
			continue
		}
		category := tx.Category
		if category == "" {
			category = "Outros"
		}
		for _, e := range tx.Entries {
			if incomeID != "" && e.AccountID == incomeID {
				income = income.Add(e.Amount)
			}
			if expenseID != "" && e.AccountID == expenseID {
				expense = expense.Add(e.Amount)
				if _, seen := categoryTotals[category]; !seen {
					categoryOrder = append(categoryOrder, category)
				}
				categoryTotals[category] = categoryTotals[category].Add(e.Amount)
			}
		}
	}

	breakdown := make([]dto.CategoryAmount, 0, len(categoryOrder))
	for _, c := range categoryOrder {
		breakdown = append(breakdown, dto.CategoryAmount{Category: c, Amount: categoryTotals[c]})
	}
	return dto.MonthlySummaryResponse{
		MonthlyIncome:     income,
		MonthlyExpense:    expense,
		CategoryBreakdown: breakdown,
	}, nil
}

// GetPlannedVsActual cross-references budget limits with the month's actual
// per-category expense.
func (s *DashboardService) GetPlannedVsActual(ctx context.Context, userID string, year, month int) ([]dto.PlannedVsActualItem, error) {
	monthly, err := s.GetMonthlySummary(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}
	budgets, err := s.budgets.ListBudgets(ctx, userID)
	if err != nil {
		return nil, err
	}

	actualByCategory := map[string]decimal.Decimal{}
	categoryOrder := []string{}
	for _, c := range monthly.CategoryBreakdown {
		actualByCategory[c.Category] = c.Amount
		categoryOrder = append(categoryOrder, c.Category)
	}

	out := make([]dto.PlannedVsActualItem, 0, len(budgets)+len(monthly.CategoryBreakdown))
	budgeted := map[string]bool{}
	for _, b := range budgets {
		out = append(out, dto.PlannedVsActualItem{
			Category: b.Category,
			Planned:  b.LimitValue,
			Actual:   actualByCategory[b.Category],
		})
		budgeted[b.Category] = true
	}
	for _, c := range categoryOrder {
		if budgeted[c] {
			continue
		}
		out = append(out, dto.PlannedVsActualItem{
			Category: c,
			Planned:  decimal.Zero,
			Actual:   actualByCategory[c],
		})
	}
	return out, nil
}

// GetBalanceTrend returns N daily end-of-day total-balance points ending today (UTC).
// Strategy: take the current total (which already reflects every transaction),
// then walk backwards subtracting each day's net change on user accounts.
func (s *DashboardService) GetBalanceTrend(ctx context.Context, userID string, days int) ([]dto.BalanceTrendPoint, error) {
	if days < 1 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	userAccounts, err := s.accounts.ListFiltered(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	userAccountIDs := make(map[string]bool, len(userAccounts))
	var currentTotal decimal.Decimal
	for _, a := range userAccounts {
		currentTotal = currentTotal.Add(a.Balance)
		userAccountIDs[a.ID] = true
	}

	txns, err := s.transactions.List(ctx, userID, 500, "")
	if err != nil {
		return nil, err
	}

	netByDay := map[string]decimal.Decimal{}
	for _, tx := range txns {
		ts, err := time.Parse(time.RFC3339Nano, tx.Timestamp)
		if err != nil {
			continue
		}
		day := ts.UTC().Format("2006-01-02")
		for _, e := range tx.Entries {
			if !userAccountIDs[e.AccountID] {
				continue
			}
			delta := e.Amount
			if e.Type == credit {
				delta = delta.Neg()
			}
			netByDay[day] = netByDay[day].Add(delta)
		}
	}

	return computeBalanceTrend(currentTotal, netByDay, time.Now().UTC(), days), nil
}

// computeBalanceTrend is the pure aggregation step. Kept separate so it can
// be unit tested without mocking Dynamo.
func computeBalanceTrend(currentTotal decimal.Decimal, netByDay map[string]decimal.Decimal, today time.Time, days int) []dto.BalanceTrendPoint {
	series := make([]dto.BalanceTrendPoint, days)
	running := currentTotal
	series[days-1] = dto.BalanceTrendPoint{
		Date:         today.Format("2006-01-02"),
		TotalBalance: running,
	}
	for i := 1; i < days; i++ {
		// To go from EOD[D] to EOD[D-1], subtract day D's net change.
		subtractDay := today.AddDate(0, 0, -(i - 1)).Format("2006-01-02")
		running = running.Sub(netByDay[subtractDay])
		series[days-1-i] = dto.BalanceTrendPoint{
			Date:         today.AddDate(0, 0, -i).Format("2006-01-02"),
			TotalBalance: running,
		}
	}
	return series
}

// GetProjection returns the end-of-month projection for the requested month.
// For past months it returns the actuals; for future months, currentBalance.
func (s *DashboardService) GetProjection(ctx context.Context, userID string, year, month int) (dto.ProjectionResponse, error) {
	summary, err := s.GetSummary(ctx, userID)
	if err != nil {
		return dto.ProjectionResponse{}, err
	}
	monthly, err := s.GetMonthlySummary(ctx, userID, year, month)
	if err != nil {
		return dto.ProjectionResponse{}, err
	}
	salary, _ := s.budgets.GetSalary(ctx, userID) // ignore errors — salary is optional

	return computeProjection(
		summary.TotalBalance, monthly.MonthlyIncome, monthly.MonthlyExpense,
		salary, time.Now().UTC(), year, month,
	), nil
}

// computeProjection is the pure projection step. month is 0-based to match
// the JS Date contract used everywhere on the dashboard side.
func computeProjection(currentBalance, currentIncome, currentExpense decimal.Decimal,
	salary *dto.SalaryResponse, now time.Time, requestedYear, requestedMonth int) dto.ProjectionResponse {

	requestedStart := time.Date(requestedYear, time.Month(requestedMonth+1), 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := requestedStart.AddDate(0, 1, -1).Day()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	switch {
	case requestedStart.Before(currentMonthStart):
		// Past month — actuals are final.
		return dto.ProjectionResponse{
			ProjectedBalance: currentBalance,
			ProjectedIncome:  currentIncome,
			ProjectedExpense: currentExpense,
			DaysRemaining:    0,
			Basis:            "past",
		}
	case requestedStart.After(currentMonthStart):
		// Future month — nothing to project yet.
		return dto.ProjectionResponse{
			ProjectedBalance: currentBalance,
			ProjectedIncome:  decimal.Zero,
			ProjectedExpense: decimal.Zero,
			DaysRemaining:    daysInMonth,
			Basis:            "future",
		}
	}

	// Current month: linear extrapolation + optional missing salary.
	daysElapsed := now.Day()
	daysRemaining := daysInMonth - daysElapsed
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	projectedExpense := extrapolateLinear(currentExpense, daysElapsed, daysInMonth)
	projectedIncome := currentIncome
	basis := "linear"

	if salary != nil && salary.DayOfMonth != nil && *salary.DayOfMonth > daysElapsed && *salary.DayOfMonth <= daysInMonth {
		// Salary not yet received this month — add it once.
		projectedIncome = currentIncome.Add(salary.Amount)
		basis = "linear-with-salary"
	} else {
		projectedIncome = extrapolateLinear(currentIncome, daysElapsed, daysInMonth)
	}

	futureIncome := projectedIncome.Sub(currentIncome)
	futureExpense := projectedExpense.Sub(currentExpense)
	projectedBalance := currentBalance.Add(futureIncome).Sub(futureExpense)

	return dto.ProjectionResponse{
		ProjectedBalance: projectedBalance,
		ProjectedIncome:  projectedIncome,
		ProjectedExpense: projectedExpense,
		DaysRemaining:    daysRemaining,
		Basis:            basis,
	}
}

var ptMonthNames = [12]string{"Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul", "Ago", "Set", "Out", "Nov", "Dez"}

// GetBudgetProjection returns a month-by-month balance projection using the
// user's configured budget limits and salary. months capped to [1, 12].
func (s *DashboardService) GetBudgetProjection(ctx context.Context, userID string, months int) (dto.BudgetProjectionResponse, error) {
	if months < 1 {
		months = 6
	}
	if months > 12 {
		months = 12
	}

	summary, err := s.GetSummary(ctx, userID)
	if err != nil {
		return dto.BudgetProjectionResponse{}, err
	}

	budgets, err := s.budgets.ListBudgets(ctx, userID)
	if err != nil {
		return dto.BudgetProjectionResponse{}, err
	}

	var totalBudget decimal.Decimal
	for _, b := range budgets {
		totalBudget = totalBudget.Add(b.LimitValue)
	}

	salary, _ := s.budgets.GetSalary(ctx, userID)
	var monthlySalary decimal.Decimal
	if salary != nil {
		monthlySalary = salary.Amount
	}

	now := time.Now().UTC()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	daysInCurrentMonth := currentMonthStart.AddDate(0, 1, 0).AddDate(0, 0, -1).Day()
	daysRemaining := daysInCurrentMonth - now.Day()
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// Current month actuals (month is 0-based)
	monthly, err := s.GetMonthlySummary(ctx, userID, now.Year(), int(now.Month())-1)
	if err != nil {
		return dto.BudgetProjectionResponse{}, err
	}

	results := make([]dto.ProjectionMonthResult, months)
	runningBalance := summary.TotalBalance

	for i := 0; i < months; i++ {
		t := now.AddDate(0, i, 0)
		year := t.Year()
		month0 := int(t.Month()) - 1 // 0-based
		isCurrent := i == 0

		var projIncome, projExpense decimal.Decimal

		if isCurrent {
			// Only project the remaining portion of this month.
			if salary != nil && salary.DayOfMonth != nil && *salary.DayOfMonth > now.Day() {
				projIncome = monthlySalary
			}
			remaining := totalBudget.Sub(monthly.MonthlyExpense)
			if remaining.IsPositive() {
				projExpense = remaining
			}
		} else {
			projIncome = monthlySalary
			projExpense = totalBudget
		}

		delta := projIncome.Sub(projExpense)
		runningBalance = runningBalance.Add(delta)

		results[i] = dto.ProjectionMonthResult{
			Year:             year,
			Month:            month0,
			Label:            fmt.Sprintf("%s/%d", ptMonthNames[month0], year),
			ProjectedIncome:  projIncome,
			ProjectedExpense: projExpense,
			Delta:            delta,
			RunningBalance:   runningBalance,
			IsCurrent:        isCurrent,
		}
	}

	return dto.BudgetProjectionResponse{
		CurrentBalance: summary.TotalBalance,
		MonthlyBudget:  totalBudget,
		MonthlySalary:  monthlySalary,
		HasBudgets:     len(budgets) > 0,
		HasSalary:      salary != nil,
		DaysRemaining:  daysRemaining,
		Months:         results,
	}, nil
}

// GetDailyExpenses returns per-day expense totals for the requested month.
// Aggregates entries on the system expense account (DEBIT side) over txns
// whose timestamp falls inside the month (UTC).
func (s *DashboardService) GetDailyExpenses(ctx context.Context, userID string, year, month int) (dto.DailyExpensesResponse, error) {
	accounts, err := s.accounts.ListFiltered(ctx, userID, true)
	if err != nil {
		return dto.DailyExpensesResponse{}, err
	}
	_, expenseID := systemAccountIDs(accounts)

	txns, err := s.transactions.List(ctx, userID, 500, "")
	if err != nil {
		return dto.DailyExpensesResponse{}, err
	}
	return computeDailyExpenses(txns, expenseID, year, month), nil
}

// computeDailyExpenses is the pure aggregation step.
func computeDailyExpenses(txns []dto.TransactionListItem, expenseAccountID string, year, month int) dto.DailyExpensesResponse {
	start := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	daysInMonth := end.AddDate(0, 0, -1).Day()

	resp := dto.DailyExpensesResponse{
		Year:        year,
		Month:       month,
		DaysInMonth: daysInMonth,
		DayTotals:   make([]decimal.Decimal, daysInMonth),
	}
	if expenseAccountID == "" {
		return resp
	}

	for _, tx := range txns {
		ts, err := time.Parse(time.RFC3339Nano, tx.Timestamp)
		if err != nil {
			continue
		}
		if ts.Before(start) || !ts.Before(end) {
			continue
		}
		hitExpense := false
		for _, e := range tx.Entries {
			if e.AccountID != expenseAccountID || e.Type != debit {
				continue
			}
			idx := ts.UTC().Day() - 1
			resp.DayTotals[idx] = resp.DayTotals[idx].Add(e.Amount)
			hitExpense = true
		}
		if hitExpense {
			resp.TransactionCount++
		}
	}
	return resp
}

// extrapolateLinear projects a monthly figure: (current / elapsed) * total.
// Guards against div-by-zero (returns current) and capping at the actual when
// the month is over.
func extrapolateLinear(current decimal.Decimal, daysElapsed, daysInMonth int) decimal.Decimal {
	if daysElapsed <= 0 || daysElapsed >= daysInMonth {
		return current
	}
	return current.Div(decimal.NewFromInt(int64(daysElapsed))).Mul(decimal.NewFromInt(int64(daysInMonth)))
}

func systemAccountIDs(accounts []dto.AccountResponse) (incomeID, expenseID string) {
	for _, a := range accounts {
		if !a.IsSystem {
			continue
		}
		switch a.Type {
		case IncomeAccountType:
			incomeID = a.ID
		case ExpenseAccountType:
			expenseID = a.ID
		}
	}
	return
}
