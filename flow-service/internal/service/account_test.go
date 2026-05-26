package service

import (
	"context"
	"errors"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

func TestAccount_Create_BasicAccount_PersistsAndReturns(t *testing.T) {
	var puts []map[string]types.AttributeValue
	fake := &fakeDynamo{
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			puts = append(puts, in.Item)
			return &awsdynamodb.PutItemOutput{}, nil
		},
	}
	svc := NewAccountService(fake, "flow-table")

	acc, err := svc.Create(context.Background(), "user-1", dto.CreateAccountRequest{
		Name:           "Nubank",
		InitialBalance: decimal.RequireFromString("100.00"),
		Color:          "#820ad1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Name != "Nubank" || !acc.Balance.Equal(decimal.RequireFromString("100.00")) || acc.Type != "ASSET" {
		t.Errorf("unexpected account: %+v", acc)
	}
	if len(puts) != 2 {
		t.Fatalf("expected 2 PutItems, got %d", len(puts))
	}
	if flowdynamo.Str(puts[0], "code") != "NUBANK" {
		t.Errorf("unexpected code: %s", flowdynamo.Str(puts[0], "code"))
	}
}

func TestAccount_Create_InvestmentAccount_StoresAnnualRate(t *testing.T) {
	rate := decimal.RequireFromString("12.5")
	var captured map[string]types.AttributeValue
	fake := &fakeDynamo{
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			if captured == nil {
				captured = in.Item
			}
			return &awsdynamodb.PutItemOutput{}, nil
		},
	}
	svc := NewAccountService(fake, "flow-table")

	acc, err := svc.Create(context.Background(), "user-1", dto.CreateAccountRequest{
		Name: "CDB", InitialBalance: decimal.RequireFromString("5000"), Investment: true, AnnualRate: &rate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Type != "INVESTMENT" {
		t.Errorf("expected INVESTMENT, got %s", acc.Type)
	}
	if flowdynamo.Num(captured, "annualRate") != "12.5" {
		t.Errorf("annualRate not persisted: %s", flowdynamo.Num(captured, "annualRate"))
	}
}

func TestAccount_ListFiltered_HidesSystemWhenAsked(t *testing.T) {
	fake := &fakeDynamo{
		QueryFunc: func(ctx context.Context, in *awsdynamodb.QueryInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.QueryOutput, error) {
			return &awsdynamodb.QueryOutput{Items: []map[string]types.AttributeValue{
				{
					"id": flowdynamo.S("acc-1"), "code": flowdynamo.S("NUBANK"), "name": flowdynamo.S("Nubank"),
					"accountType": flowdynamo.S("ASSET"), "balance": flowdynamo.N("100"), "color": flowdynamo.S("#000"),
					"isSystem": flowdynamo.Bool(false),
				},
				{
					"id": flowdynamo.S("sys-1"), "code": flowdynamo.S("SYS"), "name": flowdynamo.S("Entrada"),
					"accountType": flowdynamo.S("INCOME"), "balance": flowdynamo.N("0"), "color": flowdynamo.S("#0f0"),
					"isSystem": flowdynamo.Bool(true),
				},
			}}, nil
		},
	}
	svc := NewAccountService(fake, "flow-table")

	visible, err := svc.ListFiltered(context.Background(), "user-1", false)
	if err != nil || len(visible) != 1 || visible[0].ID != "acc-1" {
		t.Fatalf("expected 1 non-system account, got %+v err=%v", visible, err)
	}
	all, _ := svc.ListFiltered(context.Background(), "user-1", true)
	if len(all) != 2 {
		t.Errorf("expected 2 with includeSystem, got %d", len(all))
	}
}

func TestAccount_Update_SystemAccount_Rejected(t *testing.T) {
	fake := &fakeDynamo{
		GetItemFunc: func(ctx context.Context, in *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
			return &awsdynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
				"id": flowdynamo.S("sys-1"), "code": flowdynamo.S("SYS"), "name": flowdynamo.S("Saída"),
				"accountType": flowdynamo.S("EXPENSE"), "balance": flowdynamo.N("0"), "color": flowdynamo.S("#f00"),
				"isSystem": flowdynamo.Bool(true),
			}}, nil
		},
	}
	svc := NewAccountService(fake, "flow-table")

	_, err := svc.Update(context.Background(), "user-1", "sys-1", dto.UpdateAccountRequest{Name: "Hack"})
	if !errors.Is(err, ErrSystemAccountReadOnly) {
		t.Errorf("expected ErrSystemAccountReadOnly, got %v", err)
	}
}

func TestAccount_Delete_RemovesAllItemsUnderPK(t *testing.T) {
	deletes := 0
	fake := &fakeDynamo{
		QueryFunc: func(ctx context.Context, in *awsdynamodb.QueryInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.QueryOutput, error) {
			return &awsdynamodb.QueryOutput{Items: []map[string]types.AttributeValue{
				{"SK": flowdynamo.S("METADATA")},
				{"SK": flowdynamo.S("BALANCE#LATEST")},
			}}, nil
		},
		DeleteItemFunc: func(ctx context.Context, in *awsdynamodb.DeleteItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.DeleteItemOutput, error) {
			deletes++
			return &awsdynamodb.DeleteItemOutput{}, nil
		},
	}
	svc := NewAccountService(fake, "flow-table")

	if err := svc.Delete(context.Background(), "user-1", "acc-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletes != 2 {
		t.Errorf("expected 2 DeleteItem calls, got %d", deletes)
	}
}
