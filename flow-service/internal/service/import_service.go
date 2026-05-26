package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"
	"unicode"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

// ImportService handles CSV parsing and merchant rule persistence.
type ImportService struct {
	dynamo       flowdynamo.API
	tableName    string
	transactions *TransactionService
}

// NewImportService wires the service.
func NewImportService(dynamo flowdynamo.API, tableName string, transactions *TransactionService) *ImportService {
	return &ImportService{dynamo: dynamo, tableName: tableName, transactions: transactions}
}

// ParseCSV parses raw CSV bytes and applies known merchant rules.
// Returns the preview rows and the full set of persisted rules for the user.
func (s *ImportService) ParseCSV(ctx context.Context, userID string, content []byte) (dto.ImportPreviewResponse, error) {
	rules, err := s.GetMerchantRules(ctx, userID)
	if err != nil {
		return dto.ImportPreviewResponse{}, err
	}
	ruleMap := make(map[string]string, len(rules))
	for _, r := range rules {
		ruleMap[r.MerchantKey] = r.Category
	}

	rows, err := parseCSVBytes(content, ruleMap)
	if err != nil {
		return dto.ImportPreviewResponse{}, err
	}
	return dto.ImportPreviewResponse{Rows: rows, KnownRules: rules}, nil
}

// GetMerchantRules returns all saved merchant→category rules for the user.
func (s *ImportService) GetMerchantRules(ctx context.Context, userID string) ([]dto.MerchantRule, error) {
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.tableName,
		KeyConditionExpression: strPtr("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     flowdynamo.S(flowdynamo.UserPK(userID)),
			":prefix": flowdynamo.S(flowdynamo.SKMerchantRulePrefix),
		},
	})
	if err != nil {
		return nil, err
	}

	rules := make([]dto.MerchantRule, 0, len(out.Items))
	for _, item := range out.Items {
		rules = append(rules, dto.MerchantRule{
			MerchantKey: flowdynamo.Str(item, "merchantKey"),
			DisplayName: flowdynamo.Str(item, "displayName"),
			Category:    flowdynamo.Str(item, "category"),
		})
	}
	return rules, nil
}

// SaveMerchantRules upserts merchant→category rules for the user.
func (s *ImportService) SaveMerchantRules(ctx context.Context, userID string, rules []dto.MerchantRule) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, r := range rules {
		if r.MerchantKey == "" || r.Category == "" {
			continue
		}
		_, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
			TableName: &s.tableName,
			Item: map[string]types.AttributeValue{
				"PK":          flowdynamo.S(flowdynamo.UserPK(userID)),
				"SK":          flowdynamo.S(flowdynamo.MerchantRuleSK(r.MerchantKey)),
				"type":        flowdynamo.S("MERCHANT_RULE"),
				"merchantKey": flowdynamo.S(r.MerchantKey),
				"displayName": flowdynamo.S(r.DisplayName),
				"category":    flowdynamo.S(r.Category),
				"updatedAt":   flowdynamo.S(now),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Commit saves new merchant rules and posts all rows as transactions.
func (s *ImportService) Commit(ctx context.Context, userID string, req dto.ImportCommitRequest) (dto.ImportCommitResponse, error) {
	if err := s.SaveMerchantRules(ctx, userID, req.MerchantRules); err != nil {
		return dto.ImportCommitResponse{}, err
	}

	var imported, skipped int
	for _, row := range req.Rows {
		if row.Category == "" {
			skipped++
			continue
		}
		txReq := dto.PostTransactionRequest{
			Description: row.Description,
			Category:    row.Category,
			Entries: []dto.EntryRequest{
				{AccountID: req.AccountID, Amount: row.Amount, Type: row.Type},
			},
		}
		if row.Date != "" {
			txReq.ReferenceID = row.Date
		}
		if _, err := s.transactions.Post(ctx, userID, txReq); err != nil {
			return dto.ImportCommitResponse{Imported: imported, Skipped: skipped},
				fmt.Errorf("posting %q: %w", row.Description, err)
		}
		imported++
	}
	return dto.ImportCommitResponse{Imported: imported, Skipped: skipped}, nil
}

// ---------- CSV parsing ----------

// csvColIndex scans the header row for any of the given aliases (case-insensitive).
func csvColIndex(headers []string, aliases ...string) int {
	for i, h := range headers {
		norm := strings.ToLower(strings.TrimSpace(h))
		for _, alias := range aliases {
			if norm == alias {
				return i
			}
		}
	}
	return -1
}

// detectSeparator returns ',' or ';' by counting occurrences in the first line.
func detectSeparator(content []byte) rune {
	first := string(bytes.SplitN(content, []byte("\n"), 2)[0])
	if strings.Count(first, ";") >= strings.Count(first, ",") {
		return ';'
	}
	return ','
}

// parseDate accepts YYYY-MM-DD and DD/MM/YYYY (Brazilian) formats.
func parseDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) == 10 && raw[4] == '-' {
		return raw // already YYYY-MM-DD
	}
	// DD/MM/YYYY or DD/MM/YY
	parts := strings.Split(raw, "/")
	if len(parts) == 3 {
		d, m, y := parts[0], parts[1], parts[2]
		if len(y) == 2 {
			y = "20" + y
		}
		return fmt.Sprintf("%s-%s-%s", y, m, d)
	}
	return raw
}

// parseAmount normalises Brazilian (1.234,56) and US (1234.56) decimal strings.
func parseAmount(raw string) (decimal.Decimal, error) {
	s := strings.TrimSpace(raw)
	s = strings.ReplaceAll(s, " ", "")
	// Remove currency symbols
	s = strings.Trim(s, "R$€$")
	s = strings.TrimSpace(s)

	// Brazilian format: dots as thousands separator, comma as decimal
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	return decimal.NewFromString(s)
}

// toMerchantKey normalises a description into a stable lookup key.
func toMerchantKey(description string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(description) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	key := strings.Join(strings.Fields(b.String()), " ")
	if len(key) > 60 {
		key = key[:60]
	}
	return strings.TrimSpace(key)
}

func parseCSVBytes(content []byte, ruleMap map[string]string) ([]dto.ParsedRow, error) {
	sep := detectSeparator(content)

	r := csv.NewReader(bytes.NewReader(content))
	r.Comma = sep
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have a header row and at least one data row")
	}

	headers := records[0]
	dateCol := csvColIndex(headers, "data", "date", "data lançamento", "data pagamento", "data compra")
	descCol := csvColIndex(headers, "título", "titulo", "descrição", "descricao", "histórico", "historico", "description", "memo", "estabelecimento")
	amtCol := csvColIndex(headers, "valor", "value", "amount", "quantia")
	catCol := csvColIndex(headers, "categoria", "category") // optional, present in Nubank CC

	if dateCol == -1 || descCol == -1 || amtCol == -1 {
		return nil, fmt.Errorf("CSV must have date, description and amount columns (got: %s)", strings.Join(headers, ", "))
	}

	rows := make([]dto.ParsedRow, 0, len(records)-1)
	for _, rec := range records[1:] {
		if len(rec) <= amtCol || len(rec) <= descCol || len(rec) <= dateCol {
			continue
		}
		desc := strings.TrimSpace(rec[descCol])
		if desc == "" {
			continue
		}

		amt, err := parseAmount(rec[amtCol])
		if err != nil {
			continue // skip unparseable rows silently
		}

		// Determine direction: negative value = expense (DEBIT from account perspective)
		txType := "DEBIT" // expense
		if amt.Sign() > 0 {
			txType = "CREDIT" // income
		}
		amt = amt.Abs()

		date := parseDate(rec[dateCol])

		// Category from the CSV itself (e.g. Nubank CC has a Categoria column)
		var csvCategory string
		if catCol >= 0 && catCol < len(rec) {
			csvCategory = strings.TrimSpace(rec[catCol])
		}

		merchantKey := toMerchantKey(desc)
		category := ruleMap[merchantKey]
		if category == "" {
			category = csvCategory // use CSV category as fallback
		}

		rows = append(rows, dto.ParsedRow{
			Date:          date,
			Description:   desc,
			Amount:        amt,
			Type:          txType,
			Category:      category,
			NeedsCategory: category == "",
			MerchantKey:   merchantKey,
		})
	}
	return rows, nil
}
