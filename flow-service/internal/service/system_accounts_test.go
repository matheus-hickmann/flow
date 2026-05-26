package service

import (
	"context"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

func TestSystemAccounts_Ensure_NothingExists_CreatesBoth(t *testing.T) {
	puts := 0
	fake := &fakeDynamo{
		QueryFunc: func(ctx context.Context, in *awsdynamodb.QueryInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.QueryOutput, error) {
			return &awsdynamodb.QueryOutput{Items: nil}, nil
		},
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			puts++
			return &awsdynamodb.PutItemOutput{}, nil
		},
	}
	accounts := NewAccountService(fake, "flow-table")
	svc := NewSystemAccountsService(fake, "flow-table", accounts)

	sys, err := svc.Ensure(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sys.IncomeAccountID == "" || sys.ExpenseAccountID == "" || sys.IncomeAccountID == sys.ExpenseAccountID {
		t.Errorf("unexpected ids: %+v", sys)
	}
	// Two account-creates × two items each (metadata + balance snapshot) = 4 puts.
	if puts != 4 {
		t.Errorf("expected 4 PutItem calls, got %d", puts)
	}
}

func TestSystemAccounts_Ensure_BothExist_ReusesIDs(t *testing.T) {
	fake := &fakeDynamo{
		QueryFunc: func(ctx context.Context, in *awsdynamodb.QueryInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.QueryOutput, error) {
			return &awsdynamodb.QueryOutput{Items: []map[string]types.AttributeValue{
				{
					"id": flowdynamo.S("inc-id"), "code": flowdynamo.S("SYSTEM_INCOME"), "name": flowdynamo.S("Entrada"),
					"accountType": flowdynamo.S("INCOME"), "balance": flowdynamo.N("0"), "color": flowdynamo.S("#0f0"),
					"isSystem": flowdynamo.Bool(true),
				},
				{
					"id": flowdynamo.S("exp-id"), "code": flowdynamo.S("SYSTEM_EXPENSE"), "name": flowdynamo.S("Saída"),
					"accountType": flowdynamo.S("EXPENSE"), "balance": flowdynamo.N("0"), "color": flowdynamo.S("#f00"),
					"isSystem": flowdynamo.Bool(true),
				},
			}}, nil
		},
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			t.Fatal("PutItem should not be called when accounts already exist")
			return nil, nil
		},
	}
	accounts := NewAccountService(fake, "flow-table")
	svc := NewSystemAccountsService(fake, "flow-table", accounts)

	sys, err := svc.Ensure(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sys.IncomeAccountID != "inc-id" || sys.ExpenseAccountID != "exp-id" {
		t.Errorf("unexpected: %+v", sys)
	}
}
