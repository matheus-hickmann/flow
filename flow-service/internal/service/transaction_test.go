package service

import (
	"context"
	"errors"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
)

// helper: builds a TransactionService whose system-accounts dependency
// always returns the provided IDs without hitting Dynamo for the lookup.
func newTxnSvcWithSystem(fake *fakeDynamo, sys SystemAccounts) *TransactionService {
	accounts := NewAccountService(fake, "flow-table")
	// Stub the system-accounts service by replacing Ensure: build a wrapper.
	systemSvc := &SystemAccountsService{dynamo: fake, tableName: "flow-table", accounts: accounts}
	// Pre-seed the system accounts in fake's Query to avoid creation.
	fake.QueryFunc = func(ctx context.Context, in *awsdynamodb.QueryInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.QueryOutput, error) {
		if in.IndexName != nil && *in.IndexName == "GSI1" {
			return &awsdynamodb.QueryOutput{Items: []map[string]types.AttributeValue{
				stubSystemAccount(sys.IncomeAccountID, IncomeAccountType, IncomeAccountName),
				stubSystemAccount(sys.ExpenseAccountID, ExpenseAccountType, ExpenseAccountName),
			}}, nil
		}
		// fall-through: entry fetches for List() return empty
		return &awsdynamodb.QueryOutput{Items: nil}, nil
	}
	return NewTransactionService(fake, "flow-table", systemSvc)
}

func stubSystemAccount(id, accountType, name string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: id},
		"code":        &types.AttributeValueMemberS{Value: "SYSTEM_" + accountType},
		"name":        &types.AttributeValueMemberS{Value: name},
		"accountType": &types.AttributeValueMemberS{Value: accountType},
		"balance":     &types.AttributeValueMemberN{Value: "0"},
		"color":       &types.AttributeValueMemberS{Value: "#000"},
		"isSystem":    &types.AttributeValueMemberBOOL{Value: true},
	}
}

func TestTransaction_Post_SingleCredit_AddsExpenseCounterEntry(t *testing.T) {
	var captured *awsdynamodb.TransactWriteItemsInput
	fake := &fakeDynamo{
		TransactWriteItemsFunc: func(ctx context.Context, in *awsdynamodb.TransactWriteItemsInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.TransactWriteItemsOutput, error) {
			captured = in
			return &awsdynamodb.TransactWriteItemsOutput{}, nil
		},
	}
	svc := newTxnSvcWithSystem(fake, SystemAccounts{IncomeAccountID: "inc", ExpenseAccountID: "exp"})

	resp, err := svc.Post(context.Background(), "user-1", dto.PostTransactionRequest{
		Description: "Almoço",
		Entries: []dto.EntryRequest{
			{AccountID: "user-acc", Amount: decimal.RequireFromString("25"), Type: "CREDIT"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}

	// Counter entry should hit the expense account with DEBIT.
	var counterFound bool
	for _, e := range resp.Entries {
		if e.AccountID == "exp" && e.Type == "DEBIT" && e.Amount.Equal(decimal.RequireFromString("25")) {
			counterFound = true
		}
	}
	if !counterFound {
		t.Errorf("expected counter DEBIT to expense account, got %+v", resp.Entries)
	}
	if captured == nil || len(captured.TransactItems) == 0 {
		t.Error("TransactWriteItems not invoked")
	}
}

func TestTransaction_Post_SingleDebit_AddsIncomeCounterEntry(t *testing.T) {
	fake := &fakeDynamo{}
	svc := newTxnSvcWithSystem(fake, SystemAccounts{IncomeAccountID: "inc", ExpenseAccountID: "exp"})

	resp, err := svc.Post(context.Background(), "user-1", dto.PostTransactionRequest{
		Description: "Salário",
		Entries: []dto.EntryRequest{
			{AccountID: "user-acc", Amount: decimal.RequireFromString("5000"), Type: "DEBIT"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}
	var counterFound bool
	for _, e := range resp.Entries {
		if e.AccountID == "inc" && e.Type == "CREDIT" {
			counterFound = true
		}
	}
	if !counterFound {
		t.Errorf("expected counter CREDIT to income account, got %+v", resp.Entries)
	}
}

func TestTransaction_Post_UnbalancedTwoEntries_Returns400(t *testing.T) {
	fake := &fakeDynamo{}
	svc := newTxnSvcWithSystem(fake, SystemAccounts{})

	_, err := svc.Post(context.Background(), "user-1", dto.PostTransactionRequest{
		Description: "Errado",
		Entries: []dto.EntryRequest{
			{AccountID: "a", Amount: decimal.RequireFromString("10"), Type: "CREDIT"},
			{AccountID: "b", Amount: decimal.RequireFromString("5"), Type: "DEBIT"},
		},
	})
	if !errors.Is(err, ErrInvalidTransaction) {
		t.Errorf("expected ErrInvalidTransaction, got %v", err)
	}
}

func TestTransaction_Post_TransferTwoEntries_NotExpanded(t *testing.T) {
	fake := &fakeDynamo{}
	svc := newTxnSvcWithSystem(fake, SystemAccounts{})

	resp, err := svc.Post(context.Background(), "user-1", dto.PostTransactionRequest{
		Description: "Transfer",
		Entries: []dto.EntryRequest{
			{AccountID: "from", Amount: decimal.RequireFromString("100"), Type: "CREDIT"},
			{AccountID: "to", Amount: decimal.RequireFromString("100"), Type: "DEBIT"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp.Entries))
	}
}
