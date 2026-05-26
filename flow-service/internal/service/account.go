package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

// ErrAccountNotFound is returned when an account id does not exist.
var ErrAccountNotFound = errors.New("account not found")

// ErrSystemAccountReadOnly blocks edits/deletes on Entrada/Saída.
var ErrSystemAccountReadOnly = errors.New("system accounts cannot be edited")

// AccountService groups every read+write op on accounts.
// Mirrors AccountQuery + CreateAccountCommand + UpdateAccountCommand +
// DeleteAccountCommand + AdjustBalanceCommand + BalanceQuery from the Java side.
type AccountService struct {
	dynamo    flowdynamo.API
	tableName string
}

// NewAccountService wires the service.
func NewAccountService(dynamo flowdynamo.API, tableName string) *AccountService {
	return &AccountService{dynamo: dynamo, tableName: tableName}
}

// ---------- Reads ----------

// GetByID returns the account or nil when absent.
func (s *AccountService) GetByID(ctx context.Context, userID, accountID string) (*dto.AccountResponse, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.AccountPK(userID, accountID)),
			"SK": flowdynamo.S(flowdynamo.SKMetadata),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	acc := toAccountResponse(out.Item)
	return &acc, nil
}

// ListFiltered returns accounts; includeSystem=false hides Entrada/Saída.
func (s *AccountService) ListFiltered(ctx context.Context, userID string, includeSystem bool) ([]dto.AccountResponse, error) {
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.tableName,
		IndexName:              strPtr("GSI1"),
		KeyConditionExpression: strPtr("GSI1PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": flowdynamo.S(flowdynamo.GSI1PKAccounts(userID)),
		},
	})
	if err != nil {
		return nil, err
	}
	accounts := make([]dto.AccountResponse, 0, len(out.Items))
	for _, item := range out.Items {
		a := toAccountResponse(item)
		if !includeSystem && a.IsSystem {
			continue
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

// GetBalanceSnapshot reads the BALANCE#LATEST item; nil when missing.
func (s *AccountService) GetBalanceSnapshot(ctx context.Context, userID, accountID string) (*dto.BalanceItem, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.AccountPK(userID, accountID)),
			"SK": flowdynamo.S(flowdynamo.SKBalanceLatest),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	bal, _ := decimal.NewFromString(flowdynamo.Num(out.Item, "currentBalance"))
	return &dto.BalanceItem{
		CurrentBalance: bal,
		LastUpdate:     flowdynamo.Str(out.Item, "lastUpdate"),
	}, nil
}

// ---------- Writes ----------

// Create persists a new account + its initial balance snapshot.
func (s *AccountService) Create(ctx context.Context, userID string, req dto.CreateAccountRequest) (dto.AccountResponse, error) {
	id := uuid.NewString()
	code := normalizeCode(req.Name)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pk := flowdynamo.AccountPK(userID, id)
	accountType := resolveAccountType(req)

	item := map[string]types.AttributeValue{
		"PK":         flowdynamo.S(pk),
		"SK":         flowdynamo.S(flowdynamo.SKMetadata),
		"type":       flowdynamo.S("ACCOUNT"),
		"id":         flowdynamo.S(id),
		"code":       flowdynamo.S(code),
		"name":       flowdynamo.S(req.Name),
		"accountType": flowdynamo.S(accountType),
		"balance":    flowdynamo.N(req.InitialBalance.String()),
		"color":      flowdynamo.S(req.ColorOrDefault()),
		"createdAt":  flowdynamo.S(now),
		"updatedAt":  flowdynamo.S(now),
		"GSI1PK":     flowdynamo.S(flowdynamo.GSI1PKAccounts(userID)),
		"GSI1SK":     flowdynamo.S(flowdynamo.GSI1SKAccount(id)),
		"isSystem":   flowdynamo.Bool(req.System),
		"investment": flowdynamo.Bool(req.Investment),
	}
	if req.AnnualRate != nil {
		item["annualRate"] = flowdynamo.N(req.AnnualRate.String())
	}
	if req.Brand != "" {
		item["brand"] = flowdynamo.S(req.Brand)
	}
	if req.Limit != nil {
		item["creditLimit"] = flowdynamo.N(req.Limit.String())
	}
	if req.ClosingDay != nil {
		item["closingDay"] = flowdynamo.NInt(*req.ClosingDay)
	}
	if req.DueDay != nil {
		item["dueDay"] = flowdynamo.NInt(*req.DueDay)
	}

	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{TableName: &s.tableName, Item: item}); err != nil {
		return dto.AccountResponse{}, err
	}

	balanceItem := map[string]types.AttributeValue{
		"PK":             flowdynamo.S(pk),
		"SK":             flowdynamo.S(flowdynamo.SKBalanceLatest),
		"type":           flowdynamo.S("BALANCE"),
		"currentBalance": flowdynamo.N(req.InitialBalance.String()),
		"lastUpdate":     flowdynamo.S(now),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{TableName: &s.tableName, Item: balanceItem}); err != nil {
		return dto.AccountResponse{}, err
	}

	return dto.AccountResponse{
		ID: id, Code: code, Name: req.Name, Type: accountType,
		Balance: req.InitialBalance, Color: req.ColorOrDefault(),
		IsSystem: req.System, Investment: req.Investment,
		AnnualRate: req.AnnualRate, Brand: req.Brand, CreditLimit: req.Limit,
		ClosingDay: req.ClosingDay, DueDay: req.DueDay,
	}, nil
}

// Update applies a partial patch to an account's mutable fields.
func (s *AccountService) Update(ctx context.Context, userID, accountID string, req dto.UpdateAccountRequest) (dto.AccountResponse, error) {
	existing, err := s.GetByID(ctx, userID, accountID)
	if err != nil {
		return dto.AccountResponse{}, err
	}
	if existing == nil {
		return dto.AccountResponse{}, ErrAccountNotFound
	}
	if existing.IsSystem {
		return dto.AccountResponse{}, ErrSystemAccountReadOnly
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	exprParts := []string{"updatedAt = :now"}
	values := map[string]types.AttributeValue{":now": flowdynamo.S(now)}
	names := map[string]string{}
	result := *existing

	if req.Name != "" {
		exprParts = append(exprParts, "#n = :name")
		values[":name"] = flowdynamo.S(req.Name)
		names["#n"] = "name"
		result.Name = req.Name
	}
	if req.Color != "" {
		exprParts = append(exprParts, "color = :color")
		values[":color"] = flowdynamo.S(req.Color)
		result.Color = req.Color
	}
	if req.Investment != nil {
		exprParts = append(exprParts, "investment = :investment")
		values[":investment"] = flowdynamo.Bool(*req.Investment)
		result.Investment = *req.Investment
	}
	if req.Shared != nil {
		exprParts = append(exprParts, "shared = :shared")
		values[":shared"] = flowdynamo.Bool(*req.Shared)
		result.Shared = *req.Shared
	}
	if req.AnnualRate != nil {
		exprParts = append(exprParts, "annualRate = :annualRate")
		values[":annualRate"] = flowdynamo.N(req.AnnualRate.String())
		result.AnnualRate = req.AnnualRate
	}
	if req.Brand != "" {
		exprParts = append(exprParts, "brand = :brand")
		values[":brand"] = flowdynamo.S(req.Brand)
		result.Brand = req.Brand
	}
	if req.Limit != nil {
		exprParts = append(exprParts, "creditLimit = :creditLimit")
		values[":creditLimit"] = flowdynamo.N(req.Limit.String())
		result.CreditLimit = req.Limit
	}
	if req.ClosingDay != nil {
		exprParts = append(exprParts, "closingDay = :closingDay")
		values[":closingDay"] = flowdynamo.NInt(*req.ClosingDay)
		result.ClosingDay = req.ClosingDay
	}
	if req.DueDay != nil {
		exprParts = append(exprParts, "dueDay = :dueDay")
		values[":dueDay"] = flowdynamo.NInt(*req.DueDay)
		result.DueDay = req.DueDay
	}

	input := &awsdynamodb.UpdateItemInput{
		TableName:                 &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.AccountPK(userID, accountID)),
			"SK": flowdynamo.S(flowdynamo.SKMetadata),
		},
		UpdateExpression:          strPtr("SET " + strings.Join(exprParts, ", ")),
		ExpressionAttributeValues: values,
	}
	if len(names) > 0 {
		input.ExpressionAttributeNames = names
	}
	if _, err := s.dynamo.UpdateItem(ctx, input); err != nil {
		return dto.AccountResponse{}, err
	}
	return result, nil
}

// Delete removes every item under the account's PK (METADATA + BALANCE#LATEST).
// Ledger entries from posted transactions live under a different PK and remain.
func (s *AccountService) Delete(ctx context.Context, userID, accountID string) error {
	pk := flowdynamo.AccountPK(userID, accountID)
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.tableName,
		KeyConditionExpression: strPtr("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": flowdynamo.S(pk),
		},
		ProjectionExpression: strPtr("SK"),
	})
	if err != nil {
		return err
	}
	for _, item := range out.Items {
		sk := flowdynamo.Str(item, "SK")
		if _, err := s.dynamo.DeleteItem(ctx, &awsdynamodb.DeleteItemInput{
			TableName: &s.tableName,
			Key: map[string]types.AttributeValue{
				"PK": flowdynamo.S(pk),
				"SK": flowdynamo.S(sk),
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

// AdjustBalance writes a new absolute balance to both the metadata item and the
// BALANCE#LATEST snapshot.
func (s *AccountService) AdjustBalance(ctx context.Context, userID, accountID string, req dto.AdjustBalanceRequest) (dto.AccountResponse, error) {
	existing, err := s.GetByID(ctx, userID, accountID)
	if err != nil {
		return dto.AccountResponse{}, err
	}
	if existing == nil {
		return dto.AccountResponse{}, ErrAccountNotFound
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	bal := req.NewBalance.String()

	if _, err := s.dynamo.UpdateItem(ctx, &awsdynamodb.UpdateItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.AccountPK(userID, accountID)),
			"SK": flowdynamo.S(flowdynamo.SKMetadata),
		},
		UpdateExpression: strPtr("SET balance = :bal, updatedAt = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":bal": flowdynamo.N(bal),
			":now": flowdynamo.S(now),
		},
	}); err != nil {
		return dto.AccountResponse{}, err
	}

	if _, err := s.dynamo.UpdateItem(ctx, &awsdynamodb.UpdateItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.AccountPK(userID, accountID)),
			"SK": flowdynamo.S(flowdynamo.SKBalanceLatest),
		},
		UpdateExpression:         strPtr("SET currentBalance = :bal, lastUpdate = :now, #t = :type"),
		ExpressionAttributeNames: map[string]string{"#t": "type"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":bal":  flowdynamo.N(bal),
			":now":  flowdynamo.S(now),
			":type": flowdynamo.S("BALANCE"),
		},
	}); err != nil {
		return dto.AccountResponse{}, err
	}

	result := *existing
	result.Balance = req.NewBalance
	return result, nil
}

// ---------- helpers ----------

func toAccountResponse(item map[string]types.AttributeValue) dto.AccountResponse {
	bal, _ := decimal.NewFromString(flowdynamo.Num(item, "balance"))
	resp := dto.AccountResponse{
		ID:         flowdynamo.Str(item, "id"),
		Code:       flowdynamo.Str(item, "code"),
		Name:       flowdynamo.Str(item, "name"),
		Type:       flowdynamo.Str(item, "accountType"),
		Balance:    bal,
		Color:      flowdynamo.Str(item, "color"),
		IsSystem:   flowdynamo.ReadBool(item, "isSystem"),
		Investment: flowdynamo.ReadBool(item, "investment"),
		Shared:     flowdynamo.ReadBool(item, "shared"),
	}
	if v := flowdynamo.Num(item, "annualRate"); v != "" {
		d, _ := decimal.NewFromString(v)
		resp.AnnualRate = &d
	}
	if v := flowdynamo.Str(item, "brand"); v != "" {
		resp.Brand = v
	}
	if v := flowdynamo.Num(item, "creditLimit"); v != "" {
		d, _ := decimal.NewFromString(v)
		resp.CreditLimit = &d
	}
	if v := flowdynamo.Num(item, "closingDay"); v != "" {
		var i int
		_, _ = fmt.Sscanf(v, "%d", &i)
		resp.ClosingDay = &i
	}
	if v := flowdynamo.Num(item, "dueDay"); v != "" {
		var i int
		_, _ = fmt.Sscanf(v, "%d", &i)
		resp.DueDay = &i
	}
	return resp
}

var codeAlnum = regexp.MustCompile(`[^A-Z0-9]`)
var codeUnderscores = regexp.MustCompile(`_+`)

func normalizeCode(name string) string {
	upper := codeUnderscores.ReplaceAllString(
		codeAlnum.ReplaceAllString(strings.ToUpper(strings.TrimSpace(name)), "_"), "_")
	if upper == "" {
		return "ACC"
	}
	if len(upper) > 50 {
		return upper[:50]
	}
	return upper
}

func resolveAccountType(req dto.CreateAccountRequest) string {
	if req.Brand != "" {
		return "CREDIT_CARD"
	}
	if req.Investment {
		return "INVESTMENT"
	}
	return "ASSET"
}

func strPtr(s string) *string { return &s }
