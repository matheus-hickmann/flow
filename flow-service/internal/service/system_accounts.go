package service

import (
	"context"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

// Names of the virtual system accounts that act as the counterparty for
// income/expense transactions. Must match the names the frontend looks up.
const (
	IncomeAccountName  = "Entrada"
	ExpenseAccountName = "Saída"
	IncomeAccountType  = "INCOME"
	ExpenseAccountType = "EXPENSE"
)

// SystemAccounts holds the IDs of the virtual income/expense accounts for one user.
type SystemAccounts struct {
	IncomeAccountID  string
	ExpenseAccountID string
}

// SystemAccountsService lazily provisions the per-user Entrada/Saída accounts.
type SystemAccountsService struct {
	dynamo    flowdynamo.API
	tableName string
	accounts  *AccountService
}

// NewSystemAccountsService wires the service.
func NewSystemAccountsService(dynamo flowdynamo.API, tableName string, accounts *AccountService) *SystemAccountsService {
	return &SystemAccountsService{dynamo: dynamo, tableName: tableName, accounts: accounts}
}

// Ensure returns the IDs of the user's system accounts, creating any that
// don't exist yet.
func (s *SystemAccountsService) Ensure(ctx context.Context, userID string) (SystemAccounts, error) {
	all, err := s.accounts.ListFiltered(ctx, userID, true)
	if err != nil {
		return SystemAccounts{}, err
	}

	incomeID := findSystem(all, IncomeAccountType)
	expenseID := findSystem(all, ExpenseAccountType)

	if incomeID == "" {
		id, err := s.create(ctx, userID, IncomeAccountName, IncomeAccountType, "#22c55e")
		if err != nil {
			return SystemAccounts{}, err
		}
		incomeID = id
	}
	if expenseID == "" {
		id, err := s.create(ctx, userID, ExpenseAccountName, ExpenseAccountType, "#ef4444")
		if err != nil {
			return SystemAccounts{}, err
		}
		expenseID = id
	}

	return SystemAccounts{IncomeAccountID: incomeID, ExpenseAccountID: expenseID}, nil
}

func findSystem(accounts []dto.AccountResponse, accountType string) string {
	for _, a := range accounts {
		if a.IsSystem && a.Type == accountType {
			return a.ID
		}
	}
	return ""
}

// create writes the metadata + balance items directly with the right
// accountType (skips AccountService.Create because that decides type by brand/investment).
func (s *SystemAccountsService) create(ctx context.Context, userID, name, accountType, color string) (string, error) {
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pk := flowdynamo.AccountPK(userID, id)
	code := "SYSTEM_" + accountType

	item := map[string]types.AttributeValue{
		"PK":          flowdynamo.S(pk),
		"SK":          flowdynamo.S(flowdynamo.SKMetadata),
		"type":        flowdynamo.S("ACCOUNT"),
		"id":          flowdynamo.S(id),
		"code":        flowdynamo.S(code),
		"name":        flowdynamo.S(name),
		"accountType": flowdynamo.S(accountType),
		"balance":     flowdynamo.N("0"),
		"color":       flowdynamo.S(color),
		"createdAt":   flowdynamo.S(now),
		"updatedAt":   flowdynamo.S(now),
		"GSI1PK":      flowdynamo.S(flowdynamo.GSI1PKAccounts(userID)),
		"GSI1SK":      flowdynamo.S(flowdynamo.GSI1SKAccount(id)),
		"isSystem":    flowdynamo.Bool(true),
		"investment":  flowdynamo.Bool(false),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{TableName: &s.tableName, Item: item}); err != nil {
		return "", err
	}

	balance := map[string]types.AttributeValue{
		"PK":             flowdynamo.S(pk),
		"SK":             flowdynamo.S(flowdynamo.SKBalanceLatest),
		"type":           flowdynamo.S("BALANCE"),
		"currentBalance": flowdynamo.N("0"),
		"lastUpdate":     flowdynamo.S(now),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{TableName: &s.tableName, Item: balance}); err != nil {
		return "", err
	}

	return id, nil
}
