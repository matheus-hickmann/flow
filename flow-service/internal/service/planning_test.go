package service

import (
	"context"
	"errors"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

func TestPlanning_Submit_Limit_PersistsBudget(t *testing.T) {
	var captured *awsdynamodb.PutItemInput
	fake := &fakeDynamo{
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			captured = in
			return &awsdynamodb.PutItemOutput{}, nil
		},
	}
	svc := NewPlanningService(fake, "flow-table")

	limit := decimal.RequireFromString("500")
	id, err := svc.Submit(context.Background(), "user-1", dto.PlanningSubmitRequest{
		Type: "limit", Category: "Alimentação", LimitType: "ABSOLUTE", LimitValue: &limit,
	})
	if err != nil || id == "" {
		t.Fatalf("unexpected: id=%q err=%v", id, err)
	}
	sk := flowdynamo.Str(captured.Item, "SK")
	if sk[:len(flowdynamo.SKBudgetPrefix)] != flowdynamo.SKBudgetPrefix {
		t.Errorf("expected BUDGET# SK, got %s", sk)
	}
}

func TestPlanning_Submit_Salary_PersistsSalary(t *testing.T) {
	var captured *awsdynamodb.PutItemInput
	fake := &fakeDynamo{
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			captured = in
			return &awsdynamodb.PutItemOutput{}, nil
		},
	}
	svc := NewPlanningService(fake, "flow-table")

	amount := decimal.RequireFromString("5000")
	day := 5
	id, err := svc.Submit(context.Background(), "user-1", dto.PlanningSubmitRequest{
		Type: "salary", Amount: &amount, DayOfMonth: &day, AccountID: "acc-1",
	})
	if err != nil || id != flowdynamo.SKSalary {
		t.Fatalf("unexpected: id=%q err=%v", id, err)
	}
	if flowdynamo.Num(captured.Item, "amount") != "5000" {
		t.Errorf("unexpected amount: %s", flowdynamo.Num(captured.Item, "amount"))
	}
	if flowdynamo.Num(captured.Item, "dayOfMonth") != "5" {
		t.Errorf("unexpected dayOfMonth: %s", flowdynamo.Num(captured.Item, "dayOfMonth"))
	}
}

func TestPlanning_Submit_UnknownType_Rejected(t *testing.T) {
	svc := NewPlanningService(&fakeDynamo{}, "flow-table")

	_, err := svc.Submit(context.Background(), "user-1", dto.PlanningSubmitRequest{Type: "foo"})
	if !errors.Is(err, ErrInvalidPlanning) {
		t.Errorf("expected ErrInvalidPlanning, got %v", err)
	}
}

func TestPlanning_GetSalary_Missing_ReturnsNil(t *testing.T) {
	svc := NewPlanningService(&fakeDynamo{}, "flow-table")

	got, err := svc.GetSalary(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}
