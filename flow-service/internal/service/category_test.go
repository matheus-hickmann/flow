package service

import (
	"context"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

func TestCategory_Get_NoItem_ReturnsDefaults(t *testing.T) {
	svc := NewCategoryService(&fakeDynamo{}, "flow-table")

	got, err := svc.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Expense) == 0 || len(got.Income) == 0 {
		t.Error("expected defaults to be non-empty")
	}
}

func TestCategory_Get_StoredItem_ParsesCustom(t *testing.T) {
	fake := &fakeDynamo{
		GetItemFunc: func(ctx context.Context, in *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
			return &awsdynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
				"expense": categoryListAttr([]dto.CategoryItem{{ID: "pet", Name: "Pet", Color: "#f00"}}),
				"income":  categoryListAttr([]dto.CategoryItem{{ID: "aluguel", Name: "Aluguel", Color: "#0f0"}}),
			}}, nil
		},
	}
	svc := NewCategoryService(fake, "flow-table")

	got, err := svc.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Expense) != 1 || got.Expense[0].Name != "Pet" {
		t.Errorf("unexpected expense: %+v", got.Expense)
	}
	if len(got.Income) != 1 || got.Income[0].Name != "Aluguel" {
		t.Errorf("unexpected income: %+v", got.Income)
	}
}

func TestCategory_Save_FullReplace(t *testing.T) {
	var captured *awsdynamodb.PutItemInput
	fake := &fakeDynamo{
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			captured = in
			return &awsdynamodb.PutItemOutput{}, nil
		},
	}
	svc := NewCategoryService(fake, "flow-table")

	payload := dto.CategoryList{
		Expense: []dto.CategoryItem{{ID: "food", Name: "Food", Color: "#f00"}},
		Income:  []dto.CategoryItem{{ID: "sal", Name: "Salary", Color: "#0f0"}},
	}
	got, err := svc.Save(context.Background(), "user-1", payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Expense[0].Name != "Food" || got.Income[0].Name != "Salary" {
		t.Errorf("unexpected response: %+v", got)
	}

	if captured == nil {
		t.Fatal("PutItem not called")
	}
	if flowdynamo.Str(captured.Item, "PK") != "USER#user-1" {
		t.Errorf("PK mismatch: %s", flowdynamo.Str(captured.Item, "PK"))
	}
}
