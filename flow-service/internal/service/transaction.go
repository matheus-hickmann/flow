package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

const (
	debit  = "DEBIT"
	credit = "CREDIT"
)

// ErrInvalidTransaction marks user-facing validation failures (400 BAD_REQUEST).
var ErrInvalidTransaction = errors.New("invalid transaction")

// TransactionService handles posting (write) and listing (read) transactions.
type TransactionService struct {
	dynamo    flowdynamo.API
	tableName string
	system    *SystemAccountsService
}

// NewTransactionService wires the service.
func NewTransactionService(dynamo flowdynamo.API, tableName string, system *SystemAccountsService) *TransactionService {
	return &TransactionService{dynamo: dynamo, tableName: tableName, system: system}
}

// Post writes a transaction in a TransactWriteItems batch: tx metadata + entries
// + atomic balance updates. Single-entry requests are expanded with a virtual
// counter-entry against the system Entrada/Saída account.
func (s *TransactionService) Post(ctx context.Context, userID string, req dto.PostTransactionRequest) (dto.TransactionResponse, error) {
	if req.Description == "" {
		return dto.TransactionResponse{}, fmt.Errorf("%w: description is required", ErrInvalidTransaction)
	}
	entries, err := s.expandWithCounterEntry(ctx, userID, req.Entries)
	if err != nil {
		return dto.TransactionResponse{}, err
	}
	if err := validateEntries(entries); err != nil {
		return dto.TransactionResponse{}, err
	}

	txID := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	txnPK := flowdynamo.TransactionPK(userID, txID)

	writes := []types.TransactWriteItem{txItem(s.tableName, userID, txID, txnPK, now, req)}
	entryWrites, deltas := entryItemsAndDeltas(s.tableName, txnPK, now, entries)
	writes = append(writes, entryWrites...)
	for accountID, delta := range deltas {
		writes = append(writes, balanceUpdates(s.tableName, userID, accountID, delta, now)...)
	}

	if _, err := s.dynamo.TransactWriteItems(ctx, &awsdynamodb.TransactWriteItemsInput{
		TransactItems: writes,
	}); err != nil {
		return dto.TransactionResponse{}, err
	}

	resp := dto.TransactionResponse{
		ID: txID, Description: req.Description, Timestamp: now, ReferenceID: req.ReferenceID,
		Entries: make([]dto.EntryResponse, 0, len(entries)),
	}
	for _, e := range entries {
		resp.Entries = append(resp.Entries, dto.EntryResponse{
			ID: uuid.NewString(), AccountID: e.AccountID, Amount: e.Amount, Type: e.Type,
		})
	}
	return resp, nil
}

// List returns up to `limit` recent transactions, optionally filtered to those
// that touch `accountID`.
func (s *TransactionService) List(ctx context.Context, userID string, limit int, accountID string) ([]dto.TransactionListItem, error) {
	const maxLimit = 500
	if limit < 1 {
		limit = 1
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.tableName,
		IndexName:              strPtr("GSI1"),
		KeyConditionExpression: strPtr("GSI1PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": flowdynamo.S(flowdynamo.GSI1PKTransactions(userID)),
		},
		Limit: int32Ptr(int32(limit * 2)),
	})
	if err != nil {
		return nil, err
	}

	results := make([]dto.TransactionListItem, 0, limit)
	for _, item := range out.Items {
		pk := flowdynamo.Str(item, "PK")
		txID := flowdynamo.Str(item, "id")

		entries, err := s.fetchEntries(ctx, pk)
		if err != nil {
			return nil, err
		}
		if accountID != "" && !touchesAccount(entries, accountID) {
			continue
		}

		results = append(results, dto.TransactionListItem{
			ID:          txID,
			Description: flowdynamo.Str(item, "description"),
			Timestamp:   flowdynamo.Str(item, "timestamp"),
			ReferenceID: flowdynamo.Str(item, "referenceId"),
			Category:    flowdynamo.Str(item, "category"),
			Entries:     entries,
		})
		if len(results) >= limit {
			break
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Timestamp > results[j].Timestamp // descending
	})
	return results, nil
}

// ---------- helpers ----------

func (s *TransactionService) expandWithCounterEntry(ctx context.Context, userID string, entries []dto.EntryRequest) ([]dto.EntryRequest, error) {
	if len(entries) != 1 {
		return entries, nil
	}
	system, err := s.system.Ensure(ctx, userID)
	if err != nil {
		return nil, err
	}
	single := entries[0]
	isIncome := single.Type == debit
	counterAccount := system.ExpenseAccountID
	counterType := debit
	if isIncome {
		counterAccount = system.IncomeAccountID
		counterType = credit
	}
	return []dto.EntryRequest{
		single,
		{AccountID: counterAccount, Amount: single.Amount, Type: counterType},
	}, nil
}

func validateEntries(entries []dto.EntryRequest) error {
	if len(entries) < 2 {
		return fmt.Errorf("%w: at least two entries required for double-entry", ErrInvalidTransaction)
	}
	var debits, credits decimal.Decimal
	for _, e := range entries {
		if e.AccountID == "" {
			return fmt.Errorf("%w: entry accountId is required", ErrInvalidTransaction)
		}
		if e.Amount.Sign() <= 0 {
			return fmt.Errorf("%w: entry amount must be positive", ErrInvalidTransaction)
		}
		switch e.Type {
		case debit:
			debits = debits.Add(e.Amount)
		case credit:
			credits = credits.Add(e.Amount)
		default:
			return fmt.Errorf("%w: entry type must be DEBIT or CREDIT", ErrInvalidTransaction)
		}
	}
	if !debits.Equal(credits) {
		return fmt.Errorf("%w: sum of debits must equal sum of credits", ErrInvalidTransaction)
	}
	return nil
}

func txItem(tableName, userID, txID, txnPK, now string, req dto.PostTransactionRequest) types.TransactWriteItem {
	item := map[string]types.AttributeValue{
		"PK":          flowdynamo.S(txnPK),
		"SK":          flowdynamo.S(flowdynamo.SKMetadata),
		"type":        flowdynamo.S("TRANSACTION"),
		"id":          flowdynamo.S(txID),
		"description": flowdynamo.S(req.Description),
		"referenceId": flowdynamo.S(req.ReferenceID),
		"category":    flowdynamo.S(req.Category),
		"timestamp":   flowdynamo.S(now),
		"createdAt":   flowdynamo.S(now),
		"GSI1PK":      flowdynamo.S(flowdynamo.GSI1PKTransactions(userID)),
		"GSI1SK":      flowdynamo.S(flowdynamo.GSI1SKTransaction(txID)),
	}
	if req.BudgetLimitID != "" {
		item["budgetLimitId"] = flowdynamo.S(req.BudgetLimitID)
	}
	return types.TransactWriteItem{
		Put: &types.Put{TableName: &tableName, Item: item},
	}
}

func entryItemsAndDeltas(tableName, txnPK, now string, entries []dto.EntryRequest) ([]types.TransactWriteItem, map[string]decimal.Decimal) {
	writes := make([]types.TransactWriteItem, 0, len(entries))
	deltas := map[string]decimal.Decimal{}
	for seq, e := range entries {
		writes = append(writes, types.TransactWriteItem{
			Put: &types.Put{
				TableName: &tableName,
				Item: map[string]types.AttributeValue{
					"PK":        flowdynamo.S(txnPK),
					"SK":        flowdynamo.S(flowdynamo.EntrySK(e.AccountID, seq)),
					"type":      flowdynamo.S("ENTRY"),
					"accountId": flowdynamo.S(e.AccountID),
					"amount":    flowdynamo.N(e.Amount.String()),
					"entryType": flowdynamo.S(e.Type),
					"createdAt": flowdynamo.S(now),
				},
			},
		})
		delta := e.Amount
		if e.Type == credit {
			delta = delta.Neg()
		}
		current := deltas[e.AccountID]
		deltas[e.AccountID] = current.Add(delta)
	}
	return writes, deltas
}

func balanceUpdates(tableName, userID, accountID string, delta decimal.Decimal, now string) []types.TransactWriteItem {
	accPK := flowdynamo.AccountPK(userID, accountID)
	deltaStr := delta.String()
	return []types.TransactWriteItem{
		{
			Update: &types.Update{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					"PK": flowdynamo.S(accPK), "SK": flowdynamo.S(flowdynamo.SKMetadata),
				},
				UpdateExpression: strPtr("ADD balance :delta SET updatedAt = :now"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":delta": flowdynamo.N(deltaStr),
					":now":   flowdynamo.S(now),
				},
			},
		},
		{
			Update: &types.Update{
				TableName: &tableName,
				Key: map[string]types.AttributeValue{
					"PK": flowdynamo.S(accPK), "SK": flowdynamo.S(flowdynamo.SKBalanceLatest),
				},
				UpdateExpression:         strPtr("SET currentBalance = if_not_exists(currentBalance, :zero) + :delta, lastUpdate = :now, #t = :type"),
				ExpressionAttributeNames: map[string]string{"#t": "type"},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":zero":  flowdynamo.N("0"),
					":delta": flowdynamo.N(deltaStr),
					":now":   flowdynamo.S(now),
					":type":  flowdynamo.S("BALANCE"),
				},
			},
		},
	}
}

func (s *TransactionService) fetchEntries(ctx context.Context, txnPK string) ([]dto.EntryResponse, error) {
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.tableName,
		KeyConditionExpression: strPtr("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": flowdynamo.S(txnPK),
			":sk": flowdynamo.S(flowdynamo.SKEntryPrefix),
		},
	})
	if err != nil {
		return nil, err
	}
	entries := make([]dto.EntryResponse, 0, len(out.Items))
	for _, item := range out.Items {
		amt, _ := decimal.NewFromString(flowdynamo.Num(item, "amount"))
		entries = append(entries, dto.EntryResponse{
			ID:        uuid.NewString(),
			AccountID: flowdynamo.Str(item, "accountId"),
			Amount:    amt,
			Type:      flowdynamo.Str(item, "entryType"),
		})
	}
	return entries, nil
}

func touchesAccount(entries []dto.EntryResponse, accountID string) bool {
	for _, e := range entries {
		if e.AccountID == accountID {
			return true
		}
	}
	return false
}

func int32Ptr(v int32) *int32 { return &v }
