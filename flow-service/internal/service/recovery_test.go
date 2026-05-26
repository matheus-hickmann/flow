package service

import (
	"context"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

func TestRecovery_Save_PersistsExactlyOneItemWithList(t *testing.T) {
	var captured *awsdynamodb.PutItemInput
	fake := &fakeDynamo{
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			captured = in
			return &awsdynamodb.PutItemOutput{}, nil
		},
	}
	svc := NewRecoveryService(fake, "flow-table")

	err := svc.Save(context.Background(), "matheus", []dto.RecoveryQuestionItem{
		{Question: "Q1", Answer: "A1"},
		{Question: "Q2", Answer: "A2"},
		{Question: "Q3", Answer: "A3"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("PutItem not called")
	}
	list, ok := captured.Item["items"].(*types.AttributeValueMemberL)
	if !ok {
		t.Fatal("items attribute is not a list")
	}
	if len(list.Value) != 3 {
		t.Errorf("expected 3 items, got %d", len(list.Value))
	}
}

func TestRecovery_GetQuestionsOnly_ReturnsQuestionsWithoutAnswers(t *testing.T) {
	fake := &fakeDynamo{
		GetItemFunc: func(ctx context.Context, in *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
			return &awsdynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
				"items": &types.AttributeValueMemberL{Value: []types.AttributeValue{
					&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"question": flowdynamo.S("Q1"),
						"answer":   flowdynamo.S("A1"),
					}},
					&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"question": flowdynamo.S("Q2"),
						"answer":   flowdynamo.S("A2"),
					}},
				}},
			}}, nil
		},
	}
	svc := NewRecoveryService(fake, "flow-table")

	questions, err := svc.GetQuestionsOnly(context.Background(), "matheus")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questions) != 2 || questions[0] != "Q1" || questions[1] != "Q2" {
		t.Errorf("unexpected questions: %v", questions)
	}
}

func TestRecovery_GetQuestionsOnly_Missing_ReturnsEmpty(t *testing.T) {
	svc := NewRecoveryService(&fakeDynamo{}, "flow-table")

	questions, err := svc.GetQuestionsOnly(context.Background(), "matheus")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("expected empty, got %v", questions)
	}
}
