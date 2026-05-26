package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

var (
	ErrDebtNotFound  = errors.New("debt not found")
	ErrInvalidDebt   = errors.New("invalid debt payload")
	ErrDebtOverpaid  = errors.New("payment exceeds remaining amount")
)

const (
	debtStatusActive  = "ACTIVE"
	debtStatusSettled = "SETTLED"
	debtTypeToPayStr  = "TO_PAY"
	debtTypeToReceive = "TO_RECEIVE"
)

type DebtService struct {
	dynamo    flowdynamo.API
	tableName string
}

func NewDebtService(dynamo flowdynamo.API, tableName string) *DebtService {
	return &DebtService{dynamo: dynamo, tableName: tableName}
}

// Create persists a new debt and returns its id.
func (s *DebtService) Create(ctx context.Context, userID string, req dto.DebtRequest) (string, error) {
	if req.Name == "" || req.Amount == nil || req.Amount.IsNegative() || req.Amount.IsZero() {
		return "", fmt.Errorf("%w: name and a positive amount are required", ErrInvalidDebt)
	}
	if req.Type != debtTypeToPayStr && req.Type != debtTypeToReceive {
		return "", fmt.Errorf("%w: type must be TO_PAY or TO_RECEIVE", ErrInvalidDebt)
	}

	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	item := map[string]types.AttributeValue{
		"PK":           flowdynamo.S(flowdynamo.UserDebtPK(userID)),
		"SK":           flowdynamo.S(flowdynamo.DebtSK(id)),
		"id":           flowdynamo.S(id),
		"name":         flowdynamo.S(req.Name),
		"amount":       flowdynamo.N(req.Amount.String()),
		"remaining":    flowdynamo.N(req.Amount.String()),
		"type":         flowdynamo.S(req.Type),
		"counterparty": flowdynamo.S(req.Counterparty),
		"status":       flowdynamo.S(debtStatusActive),
		"createdAt":    flowdynamo.S(now),
	}
	if req.DueDate != "" {
		item["dueDate"] = flowdynamo.S(req.DueDate)
	}
	if req.Notes != "" {
		item["notes"] = flowdynamo.S(req.Notes)
	}

	_, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      item,
	})
	return id, err
}

// List returns all debts for a user, sorted by createdAt ascending.
func (s *DebtService) List(ctx context.Context, userID string) ([]dto.DebtResponse, error) {
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.tableName,
		KeyConditionExpression: strPtr("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": flowdynamo.S(flowdynamo.UserDebtPK(userID)),
			":sk": flowdynamo.S(flowdynamo.SKDebtPrefix),
		},
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.DebtResponse, 0, len(out.Items))
	for _, item := range out.Items {
		result = append(result, itemToDebtResponse(item))
	}
	return result, nil
}

// Get returns a single debt by id.
func (s *DebtService) Get(ctx context.Context, userID, id string) (*dto.DebtResponse, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.UserDebtPK(userID)),
			"SK": flowdynamo.S(flowdynamo.DebtSK(id)),
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, ErrDebtNotFound
	}
	r := itemToDebtResponse(out.Item)
	return &r, nil
}

// RecordPayment reduces the remaining balance. If remaining reaches zero the
// debt is automatically marked SETTLED.
func (s *DebtService) RecordPayment(ctx context.Context, userID, id string, req dto.DebtPaymentRequest) (*dto.DebtResponse, error) {
	if req.Amount == nil || req.Amount.IsNegative() || req.Amount.IsZero() {
		return nil, fmt.Errorf("%w: payment amount must be positive", ErrInvalidDebt)
	}

	debt, err := s.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if debt.Status == debtStatusSettled {
		return nil, fmt.Errorf("%w: debt is already settled", ErrInvalidDebt)
	}

	newRemaining := debt.Remaining.Sub(*req.Amount)
	if newRemaining.IsNegative() {
		return nil, ErrDebtOverpaid
	}

	newStatus := debtStatusActive
	if newRemaining.IsZero() {
		newStatus = debtStatusSettled
	}

	updateExpr := "SET #r = :r, #s = :s"
	_, err = s.dynamo.UpdateItem(ctx, &awsdynamodb.UpdateItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.UserDebtPK(userID)),
			"SK": flowdynamo.S(flowdynamo.DebtSK(id)),
		},
		UpdateExpression: &updateExpr,
		ExpressionAttributeNames: map[string]string{
			"#r": "remaining",
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":r": flowdynamo.N(newRemaining.String()),
			":s": flowdynamo.S(newStatus),
		},
	})
	if err != nil {
		return nil, err
	}

	debt.Remaining = newRemaining
	debt.Status = newStatus
	return debt, nil
}

// Delete removes a debt permanently.
func (s *DebtService) Delete(ctx context.Context, userID, id string) error {
	_, err := s.dynamo.DeleteItem(ctx, &awsdynamodb.DeleteItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.UserDebtPK(userID)),
			"SK": flowdynamo.S(flowdynamo.DebtSK(id)),
		},
	})
	return err
}

func itemToDebtResponse(item map[string]types.AttributeValue) dto.DebtResponse {
	amount, _ := decimal.NewFromString(flowdynamo.Num(item, "amount"))
	remaining, _ := decimal.NewFromString(flowdynamo.Num(item, "remaining"))
	return dto.DebtResponse{
		ID:           flowdynamo.Str(item, "id"),
		Name:         flowdynamo.Str(item, "name"),
		Amount:       amount,
		Remaining:    remaining,
		Type:         flowdynamo.Str(item, "type"),
		Counterparty: flowdynamo.Str(item, "counterparty"),
		DueDate:      flowdynamo.Str(item, "dueDate"),
		Notes:        flowdynamo.Str(item, "notes"),
		Status:       flowdynamo.Str(item, "status"),
		CreatedAt:    flowdynamo.Str(item, "createdAt"),
	}
}
