package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
)

const (
	reportTypeExpense = "expense"
	reportTypeIncome  = "income"
)

// ReportService aggregates transactions by category × month for the bar-chart
// report screen.
type ReportService struct {
	accounts     *AccountService
	transactions *TransactionService
	categories   *CategoryService
}

// NewReportService wires the service.
func NewReportService(accounts *AccountService, transactions *TransactionService, categories *CategoryService) *ReportService {
	return &ReportService{accounts: accounts, transactions: transactions, categories: categories}
}

// seriesKey lets us bucket per (type, category) without colliding when the
// same category name appears on both expense and income.
type seriesKey struct{ typ, category string }

// Monthly returns one series per (type, category) over the requested period.
// from/to are "yyyy-MM" strings. type can be "expense", "income", or anything
// else (interpreted as "all").
func (s *ReportService) Monthly(ctx context.Context, userID, from, to, reportType, expenseAccountID, incomeAccountID string) (dto.MonthlyReport, error) {
	start, err := parseYearMonth(from)
	if err != nil {
		return dto.MonthlyReport{}, fmt.Errorf("invalid from: %w", err)
	}
	end, err := parseYearMonth(to)
	if err != nil {
		return dto.MonthlyReport{}, fmt.Errorf("invalid to: %w", err)
	}
	if end.Before(start) {
		end = start
	}
	months := monthsBetween(start, end)
	rangeEnd := end.AddDate(0, 1, 0)

	// Resolve system-account IDs when the caller didn't pass them.
	if expenseAccountID == "" || incomeAccountID == "" {
		accs, err := s.accounts.ListFiltered(ctx, userID, true)
		if err != nil {
			return dto.MonthlyReport{}, err
		}
		incID, expID := systemAccountIDs(accs)
		if incomeAccountID == "" {
			incomeAccountID = incID
		}
		if expenseAccountID == "" {
			expenseAccountID = expID
		}
	}

	colors, err := s.loadCategoryColors(ctx, userID)
	if err != nil {
		return dto.MonthlyReport{}, err
	}

	includeExpense := reportType != reportTypeIncome
	includeIncome := reportType != reportTypeExpense

	txns, err := s.transactions.List(ctx, userID, 500, "")
	if err != nil {
		return dto.MonthlyReport{}, err
	}

	series := map[seriesKey]*dto.CategorySeries{}
	order := []seriesKey{}

	for _, tx := range txns {
		ts, err := time.Parse(time.RFC3339Nano, tx.Timestamp)
		if err != nil {
			continue
		}
		if ts.Before(start) || !ts.Before(rangeEnd) {
			continue
		}
		month := ts.Format("2006-01")
		category := tx.Category
		if category == "" {
			category = "Outros"
		}
		for _, e := range tx.Entries {
			if includeExpense && e.AccountID == expenseAccountID && e.Type == debit {
				addBucket(series, &order, reportTypeExpense, category, month, e.Amount, colors)
			}
			if includeIncome && e.AccountID == incomeAccountID && e.Type == credit {
				addBucket(series, &order, reportTypeIncome, category, month, e.Amount, colors)
			}
		}
	}

	result := make([]dto.CategorySeries, 0, len(order))
	for _, k := range order {
		result = append(result, *series[k])
	}
	return dto.MonthlyReport{Months: months, Series: result}, nil
}

func (s *ReportService) loadCategoryColors(ctx context.Context, userID string) (map[string]string, error) {
	list, err := s.categories.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	colors := map[string]string{}
	for _, c := range list.Expense {
		colors[c.Name] = c.Color
	}
	for _, c := range list.Income {
		if _, ok := colors[c.Name]; !ok {
			colors[c.Name] = c.Color
		}
	}
	return colors, nil
}

func addBucket(series map[seriesKey]*dto.CategorySeries, order *[]seriesKey, typ, category, month string, amount decimal.Decimal, colors map[string]string) {
	key := seriesKey{typ: typ, category: category}
	bucket, ok := series[key]
	if !ok {
		bucket = &dto.CategorySeries{
			Category: category,
			Color:    colors[category],
			Type:     typ,
			ByMonth:  map[string]decimal.Decimal{},
		}
		series[key] = bucket
		*order = append(*order, key)
	}
	bucket.ByMonth[month] = bucket.ByMonth[month].Add(amount)
}

func parseYearMonth(s string) (time.Time, error) {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected yyyy-MM, got %q: %w", s, err)
	}
	return t.UTC(), nil
}

func monthsBetween(start, end time.Time) []string {
	out := []string{}
	cursor := start
	for !cursor.After(end) {
		out = append(out, cursor.Format("2006-01"))
		cursor = cursor.AddDate(0, 1, 0)
	}
	return out
}
