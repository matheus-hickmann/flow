package service

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
)

func TestComputeBalanceTrend_NoTransactions_FlatSeries(t *testing.T) {
	today := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)

	series := computeBalanceTrend(decimal.NewFromInt(1000), map[string]decimal.Decimal{}, today, 5)

	if len(series) != 5 {
		t.Fatalf("expected 5 points, got %d", len(series))
	}
	for i, p := range series {
		if !p.TotalBalance.Equal(decimal.NewFromInt(1000)) {
			t.Errorf("point %d balance=%s, want 1000 (no txns means flat)", i, p.TotalBalance)
		}
	}
	if series[len(series)-1].Date != "2026-05-24" {
		t.Errorf("last date = %s, want today 2026-05-24", series[len(series)-1].Date)
	}
	if series[0].Date != "2026-05-20" {
		t.Errorf("first date = %s, want 2026-05-20 (4 days ago)", series[0].Date)
	}
}

func TestComputeBalanceTrend_TodayExpense_OnlyTodayPointReflectsIt(t *testing.T) {
	today := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)
	// Started at 1000, expense of 25 today → currentTotal = 975
	currentTotal := decimal.NewFromInt(975)
	net := map[string]decimal.Decimal{
		"2026-05-24": decimal.NewFromInt(-25), // CREDIT on user account = negative net
	}

	series := computeBalanceTrend(currentTotal, net, today, 3)

	// today = 975, yesterday = 1000, day before = 1000
	if !series[2].TotalBalance.Equal(decimal.NewFromInt(975)) {
		t.Errorf("today balance = %s, want 975", series[2].TotalBalance)
	}
	if !series[1].TotalBalance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("yesterday balance = %s, want 1000", series[1].TotalBalance)
	}
	if !series[0].TotalBalance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("day before balance = %s, want 1000", series[0].TotalBalance)
	}
}

func TestComputeBalanceTrend_IncomeAndExpenseAcrossDays(t *testing.T) {
	today := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)
	// Story:
	//   d-3 (2026-05-21): nothing
	//   d-2 (2026-05-22): salary +5000 (DEBIT) → net +5000
	//   d-1 (2026-05-23): rent -1500 (CREDIT) → net -1500
	//   d-0 (2026-05-24): coffee -10 (CREDIT) → net -10
	// Start balance was 1000. Current = 1000 + 5000 - 1500 - 10 = 4490
	currentTotal := decimal.NewFromInt(4490)
	net := map[string]decimal.Decimal{
		"2026-05-22": decimal.NewFromInt(5000),
		"2026-05-23": decimal.NewFromInt(-1500),
		"2026-05-24": decimal.NewFromInt(-10),
	}

	series := computeBalanceTrend(currentTotal, net, today, 4)

	expect := map[string]int64{
		"2026-05-21": 1000, // eod d-3 — before any movement
		"2026-05-22": 6000, // +5000 salary
		"2026-05-23": 4500, // -1500 rent
		"2026-05-24": 4490, // -10 coffee = today
	}
	for _, p := range series {
		want := expect[p.Date]
		if !p.TotalBalance.Equal(decimal.NewFromInt(want)) {
			t.Errorf("balance on %s = %s, want %d", p.Date, p.TotalBalance, want)
		}
	}
}

func TestComputeBalanceTrend_DaysClampingInService(t *testing.T) {
	// Helper accepts any days >= 1; the clamping happens in GetBalanceTrend.
	// Sanity-check the helper handles days=1 without panic.
	series := computeBalanceTrend(decimal.NewFromInt(100), nil, time.Now().UTC(), 1)
	if len(series) != 1 {
		t.Errorf("expected 1 point, got %d", len(series))
	}
}

// ─── Projection helper ─────────────────────────────────────────────────

func TestComputeProjection_CurrentMonth_LinearExtrapolation(t *testing.T) {
	// 2026-05 has 31 days. now=2026-05-10 (day 10, so 1/3 elapsed).
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	currentBalance := decimal.NewFromInt(1000)
	currentIncome := decimal.NewFromInt(1500)
	currentExpense := decimal.NewFromInt(600)

	r := computeProjection(currentBalance, currentIncome, currentExpense, nil, now, 2026, 4) // month=4 → May

	// projectedExpense = 600 / 10 * 31 = 1860
	if !r.ProjectedExpense.Equal(decimal.NewFromInt(1860)) {
		t.Errorf("projectedExpense = %s, want 1860", r.ProjectedExpense)
	}
	// projectedIncome = 1500 / 10 * 31 = 4650
	if !r.ProjectedIncome.Equal(decimal.NewFromInt(4650)) {
		t.Errorf("projectedIncome = %s, want 4650", r.ProjectedIncome)
	}
	// projectedBalance = 1000 + (4650-1500) - (1860-600) = 1000 + 3150 - 1260 = 2890
	if !r.ProjectedBalance.Equal(decimal.NewFromInt(2890)) {
		t.Errorf("projectedBalance = %s, want 2890", r.ProjectedBalance)
	}
	if r.DaysRemaining != 21 {
		t.Errorf("daysRemaining = %d, want 21", r.DaysRemaining)
	}
	if r.Basis != "linear" {
		t.Errorf("basis = %s, want linear", r.Basis)
	}
}

func TestComputeProjection_SalaryStillAhead_OverridesIncomeProjection(t *testing.T) {
	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC) // day 3
	day := 25
	salary := &dto.SalaryResponse{Amount: decimal.NewFromInt(5000), DayOfMonth: &day}

	r := computeProjection(
		decimal.NewFromInt(2000), decimal.NewFromInt(0), decimal.NewFromInt(100),
		salary, now, 2026, 4,
	)

	if !r.ProjectedIncome.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("projectedIncome = %s, want 5000 (just the salary, no extrapolation)", r.ProjectedIncome)
	}
	if r.Basis != "linear-with-salary" {
		t.Errorf("basis = %s, want linear-with-salary", r.Basis)
	}
}

func TestComputeProjection_SalaryAlreadyPaid_DoesNotDoubleCount(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC) // day 20
	day := 5
	salary := &dto.SalaryResponse{Amount: decimal.NewFromInt(5000), DayOfMonth: &day}

	// Salary day was day 5 (past), so basis should NOT be -with-salary.
	r := computeProjection(
		decimal.NewFromInt(3000), decimal.NewFromInt(5000), decimal.NewFromInt(800),
		salary, now, 2026, 4,
	)

	if r.Basis != "linear" {
		t.Errorf("basis = %s, want plain linear (salary already received)", r.Basis)
	}
}

func TestComputeProjection_PastMonth_ReturnsActuals(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	r := computeProjection(
		decimal.NewFromInt(1000), decimal.NewFromInt(5000), decimal.NewFromInt(2000),
		nil, now, 2026, 2, // March (past)
	)
	if r.Basis != "past" {
		t.Errorf("basis = %s, want past", r.Basis)
	}
	if !r.ProjectedBalance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("projectedBalance should equal currentBalance for past month")
	}
	if r.DaysRemaining != 0 {
		t.Errorf("daysRemaining should be 0 for past month")
	}
}

func TestComputeProjection_FutureMonth_ReturnsCurrentBalance(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	r := computeProjection(
		decimal.NewFromInt(1000), decimal.NewFromInt(5000), decimal.NewFromInt(2000),
		nil, now, 2026, 7, // August (future)
	)
	if r.Basis != "future" {
		t.Errorf("basis = %s, want future", r.Basis)
	}
	if !r.ProjectedBalance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("projectedBalance should equal currentBalance for future month")
	}
	if !r.ProjectedIncome.IsZero() || !r.ProjectedExpense.IsZero() {
		t.Errorf("future month should have zero projected income/expense, got %+v", r)
	}
}

func TestComputeProjection_FirstDayOfMonth_NoExtrapolation(t *testing.T) {
	// Edge case: day 1, no data yet — should not divide by zero.
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	r := computeProjection(decimal.NewFromInt(1000), decimal.Zero, decimal.Zero, nil, now, 2026, 4)
	if !r.ProjectedExpense.IsZero() || !r.ProjectedIncome.IsZero() {
		t.Errorf("first day with no data should not extrapolate, got expense=%s income=%s", r.ProjectedExpense, r.ProjectedIncome)
	}
}

// ─── Daily-expenses helper ──────────────────────────────────────────────

func TestComputeDailyExpenses_NoTransactions_AllZeros(t *testing.T) {
	r := computeDailyExpenses(nil, "exp-id", 2026, 4) // May → 31 days

	if r.DaysInMonth != 31 {
		t.Errorf("DaysInMonth = %d, want 31 (May)", r.DaysInMonth)
	}
	if len(r.DayTotals) != 31 {
		t.Errorf("len(DayTotals) = %d, want 31", len(r.DayTotals))
	}
	for i, v := range r.DayTotals {
		if !v.IsZero() {
			t.Errorf("day %d total = %s, want 0", i+1, v)
		}
	}
	if r.TransactionCount != 0 {
		t.Errorf("TransactionCount = %d, want 0", r.TransactionCount)
	}
}

func TestComputeDailyExpenses_AggregatesByDay(t *testing.T) {
	txns := []dto.TransactionListItem{
		{
			Timestamp: "2026-05-05T10:00:00Z",
			Entries: []dto.EntryResponse{
				{AccountID: "user", Amount: decimal.NewFromInt(25), Type: "CREDIT"},
				{AccountID: "exp-id", Amount: decimal.NewFromInt(25), Type: "DEBIT"},
			},
		},
		{
			Timestamp: "2026-05-05T19:00:00Z",
			Entries: []dto.EntryResponse{
				{AccountID: "user", Amount: decimal.NewFromInt(10), Type: "CREDIT"},
				{AccountID: "exp-id", Amount: decimal.NewFromInt(10), Type: "DEBIT"},
			},
		},
		{
			Timestamp: "2026-05-12T12:00:00Z",
			Entries: []dto.EntryResponse{
				{AccountID: "user", Amount: decimal.NewFromInt(100), Type: "CREDIT"},
				{AccountID: "exp-id", Amount: decimal.NewFromInt(100), Type: "DEBIT"},
			},
		},
	}
	r := computeDailyExpenses(txns, "exp-id", 2026, 4)

	if !r.DayTotals[4].Equal(decimal.NewFromInt(35)) { // day 5 = index 4
		t.Errorf("day 5 total = %s, want 35", r.DayTotals[4])
	}
	if !r.DayTotals[11].Equal(decimal.NewFromInt(100)) { // day 12
		t.Errorf("day 12 total = %s, want 100", r.DayTotals[11])
	}
	if r.TransactionCount != 3 {
		t.Errorf("TransactionCount = %d, want 3", r.TransactionCount)
	}
}

func TestComputeDailyExpenses_IgnoresOutsideMonth(t *testing.T) {
	txns := []dto.TransactionListItem{
		{
			Timestamp: "2026-04-30T23:59:59Z",
			Entries: []dto.EntryResponse{
				{AccountID: "exp-id", Amount: decimal.NewFromInt(50), Type: "DEBIT"},
			},
		},
		{
			Timestamp: "2026-06-01T00:00:00Z",
			Entries: []dto.EntryResponse{
				{AccountID: "exp-id", Amount: decimal.NewFromInt(99), Type: "DEBIT"},
			},
		},
	}
	r := computeDailyExpenses(txns, "exp-id", 2026, 4)

	for i, v := range r.DayTotals {
		if !v.IsZero() {
			t.Errorf("day %d total = %s, want 0 (out-of-month txns must be ignored)", i+1, v)
		}
	}
}

func TestComputeDailyExpenses_IgnoresIncomeAndUserEntries(t *testing.T) {
	txns := []dto.TransactionListItem{
		{
			Timestamp: "2026-05-15T10:00:00Z",
			Entries: []dto.EntryResponse{
				{AccountID: "user", Amount: decimal.NewFromInt(5000), Type: "DEBIT"}, // income
				{AccountID: "inc-id", Amount: decimal.NewFromInt(5000), Type: "CREDIT"},
			},
		},
	}
	r := computeDailyExpenses(txns, "exp-id", 2026, 4)

	if !r.DayTotals[14].IsZero() {
		t.Errorf("income txn should not appear in expense heatmap, got %s on day 15", r.DayTotals[14])
	}
	if r.TransactionCount != 0 {
		t.Errorf("income-only txn must not bump TransactionCount, got %d", r.TransactionCount)
	}
}

func TestComputeDailyExpenses_NoSystemExpenseAccount_ReturnsEmpty(t *testing.T) {
	txns := []dto.TransactionListItem{
		{
			Timestamp: "2026-05-05T10:00:00Z",
			Entries: []dto.EntryResponse{
				{AccountID: "anything", Amount: decimal.NewFromInt(25), Type: "DEBIT"},
			},
		},
	}
	r := computeDailyExpenses(txns, "", 2026, 4) // empty expense ID

	if r.DaysInMonth != 31 || len(r.DayTotals) != 31 {
		t.Errorf("structure should still be filled, got %+v", r)
	}
	for _, v := range r.DayTotals {
		if !v.IsZero() {
			t.Errorf("expected zeros when no system expense account, got %s", v)
		}
	}
}
