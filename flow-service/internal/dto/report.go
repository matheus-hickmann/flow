package dto

import "github.com/shopspring/decimal"

// CategorySeries is one entry × month-bucket aggregation for the bar chart.
type CategorySeries struct {
	Category string                     `json:"category"`
	Color    string                     `json:"color,omitempty"`
	Type     string                     `json:"type"` // expense | income
	ByMonth  map[string]decimal.Decimal `json:"byMonth"`
}

// MonthlyReport is GET /api/v1/reports/monthly.
type MonthlyReport struct {
	Months []string         `json:"months"`
	Series []CategorySeries `json:"series"`
}
