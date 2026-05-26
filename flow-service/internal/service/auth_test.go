package service

import (
	"context"
	"errors"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

func TestAuth_Create_NewUser_PersistsAndReturns(t *testing.T) {
	var captured map[string]types.AttributeValue
	fake := &fakeDynamo{
		// First call: FindByUserID returns empty.
		GetItemFunc: func(ctx context.Context, in *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
			return &awsdynamodb.GetItemOutput{}, nil
		},
		PutItemFunc: func(ctx context.Context, in *awsdynamodb.PutItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.PutItemOutput, error) {
			captured = in.Item
			return &awsdynamodb.PutItemOutput{}, nil
		},
	}
	svc := NewAuthService(fake, "flow-table")

	user, err := svc.Create(context.Background(), dto.SignupRequest{
		UserID: "matheus", Password: "secret", DisplayName: "Matheus",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != "matheus" || user.DisplayName != "Matheus" {
		t.Errorf("unexpected user: %+v", user)
	}
	if flowdynamo.Str(captured, "PK") != "AUTH#matheus" {
		t.Errorf("PK mismatch: %s", flowdynamo.Str(captured, "PK"))
	}
	if flowdynamo.Str(captured, "displayName") != "Matheus" {
		t.Errorf("displayName not persisted")
	}
}

func TestAuth_Create_ExistingUser_ReturnsErrUserAlreadyExists(t *testing.T) {
	fake := &fakeDynamo{
		GetItemFunc: func(ctx context.Context, in *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
			return &awsdynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
				"userId": flowdynamo.S("matheus"),
			}}, nil
		},
	}
	svc := NewAuthService(fake, "flow-table")

	_, err := svc.Create(context.Background(), dto.SignupRequest{UserID: "matheus", Password: "x"})
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuth_ValidateLogin_CorrectPassword_ReturnsUser(t *testing.T) {
	fake := &fakeDynamo{
		GetItemFunc: func(ctx context.Context, in *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
			return &awsdynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
				"userId":      flowdynamo.S("matheus"),
				"password":    flowdynamo.S("secret"),
				"displayName": flowdynamo.S("Matheus"),
			}}, nil
		},
	}
	svc := NewAuthService(fake, "flow-table")

	user, err := svc.ValidateLogin(context.Background(), "matheus", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil || user.UserID != "matheus" || user.DisplayName != "Matheus" {
		t.Errorf("unexpected user: %+v", user)
	}
}

func TestAuth_ValidateLogin_WrongPassword_ReturnsNil(t *testing.T) {
	fake := &fakeDynamo{
		GetItemFunc: func(ctx context.Context, in *awsdynamodb.GetItemInput, _ ...func(*awsdynamodb.Options)) (*awsdynamodb.GetItemOutput, error) {
			return &awsdynamodb.GetItemOutput{Item: map[string]types.AttributeValue{
				"userId":   flowdynamo.S("matheus"),
				"password": flowdynamo.S("secret"),
			}}, nil
		},
	}
	svc := NewAuthService(fake, "flow-table")

	user, err := svc.ValidateLogin(context.Background(), "matheus", "wrong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %+v", user)
	}
}

func TestAuth_FindByUserID_NotFound_ReturnsNil(t *testing.T) {
	svc := NewAuthService(&fakeDynamo{}, "flow-table")

	user, err := svc.FindByUserID(context.Background(), "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Errorf("expected nil, got %+v", user)
	}
}
