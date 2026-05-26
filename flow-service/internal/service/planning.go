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

// Planning submit types match the Java enum-ish dispatch.
const (
	planTypeLimit  = "limit"
	planTypeGoal   = "goal"
	planTypeParams = "params"
	planTypeSalary = "salary"
)

// ErrInvalidPlanning is returned when the discriminator is unknown or the
// required fields for the chosen type are missing.
var ErrInvalidPlanning = errors.New("invalid planning payload")

// PlanningService stores budgets / goals / econ-params / salary under one PK
// per user (USER#{userId}#PLAN) and dispatches on type at submit time.
type PlanningService struct {
	dynamo    flowdynamo.API
	tableName string
}

// NewPlanningService wires the service.
func NewPlanningService(dynamo flowdynamo.API, tableName string) *PlanningService {
	return &PlanningService{dynamo: dynamo, tableName: tableName}
}

// Submit dispatches by type and returns the SK / id of the persisted item.
func (s *PlanningService) Submit(ctx context.Context, userID string, req dto.PlanningSubmitRequest) (string, error) {
	switch req.Type {
	case planTypeLimit:
		return s.saveBudget(ctx, userID, req)
	case planTypeGoal:
		return s.saveGoal(ctx, userID, req)
	case planTypeParams:
		return s.saveEconParams(ctx, userID, req)
	case planTypeSalary:
		return s.saveSalary(ctx, userID, req)
	default:
		return "", fmt.Errorf("%w: unknown type %q", ErrInvalidPlanning, req.Type)
	}
}

// ---------- writes ----------

func (s *PlanningService) saveBudget(ctx context.Context, userID string, req dto.PlanningSubmitRequest) (string, error) {
	if req.Category == "" || req.LimitValue == nil {
		return "", fmt.Errorf("%w: category and limitValue are required for limit", ErrInvalidPlanning)
	}
	id := uuid.NewString()
	limitType := req.LimitType
	if limitType == "" {
		limitType = "ABSOLUTE"
	}
	_, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.tableName,
		Item: map[string]types.AttributeValue{
			"PK":         flowdynamo.S(flowdynamo.UserPlanPK(userID)),
			"SK":         flowdynamo.S(flowdynamo.BudgetSK(id)),
			"type":       flowdynamo.S("BUDGET"),
			"id":         flowdynamo.S(id),
			"category":   flowdynamo.S(req.Category),
			"limitType":  flowdynamo.S(limitType),
			"limitValue": flowdynamo.N(req.LimitValue.String()),
			"createdAt":  flowdynamo.S(time.Now().UTC().Format(time.RFC3339Nano)),
		},
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *PlanningService) saveGoal(ctx context.Context, userID string, req dto.PlanningSubmitRequest) (string, error) {
	if req.Name == "" {
		return "", fmt.Errorf("%w: name is required for goal", ErrInvalidPlanning)
	}
	id := uuid.NewString()
	item := map[string]types.AttributeValue{
		"PK":        flowdynamo.S(flowdynamo.UserPlanPK(userID)),
		"SK":        flowdynamo.S(flowdynamo.GoalSK(id)),
		"type":      flowdynamo.S("GOAL"),
		"id":        flowdynamo.S(id),
		"name":      flowdynamo.S(req.Name),
		"createdAt": flowdynamo.S(time.Now().UTC().Format(time.RFC3339Nano)),
	}
	if req.ExpectedReturnRate != nil {
		item["expectedReturnRate"] = flowdynamo.N(req.ExpectedReturnRate.String())
	}
	if req.MonthlyContribution != nil {
		item["monthlyContribution"] = flowdynamo.N(req.MonthlyContribution.String())
	}
	if req.TargetAmount != nil {
		item["targetAmount"] = flowdynamo.N(req.TargetAmount.String())
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{TableName: &s.tableName, Item: item}); err != nil {
		return "", err
	}
	return id, nil
}

func (s *PlanningService) saveEconParams(ctx context.Context, userID string, req dto.PlanningSubmitRequest) (string, error) {
	if req.SelicRate == nil || req.IpcaRate == nil {
		return "", fmt.Errorf("%w: selicRate and ipcaRate are required for params", ErrInvalidPlanning)
	}
	_, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.tableName,
		Item: map[string]types.AttributeValue{
			"PK":        flowdynamo.S(flowdynamo.UserPlanPK(userID)),
			"SK":        flowdynamo.S(flowdynamo.SKEconParams),
			"type":      flowdynamo.S("ECON_PARAMS"),
			"selicRate": flowdynamo.N(req.SelicRate.String()),
			"ipcaRate":  flowdynamo.N(req.IpcaRate.String()),
			"updatedAt": flowdynamo.S(time.Now().UTC().Format(time.RFC3339Nano)),
		},
	})
	if err != nil {
		return "", err
	}
	return flowdynamo.SKEconParams, nil
}

func (s *PlanningService) saveSalary(ctx context.Context, userID string, req dto.PlanningSubmitRequest) (string, error) {
	if req.Amount == nil || req.Amount.Sign() <= 0 {
		return "", fmt.Errorf("%w: amount > 0 is required for salary", ErrInvalidPlanning)
	}
	item := map[string]types.AttributeValue{
		"PK":        flowdynamo.S(flowdynamo.UserPlanPK(userID)),
		"SK":        flowdynamo.S(flowdynamo.SKSalary),
		"type":      flowdynamo.S("SALARY"),
		"amount":    flowdynamo.N(req.Amount.String()),
		"updatedAt": flowdynamo.S(time.Now().UTC().Format(time.RFC3339Nano)),
	}
	if req.DayOfMonth != nil {
		item["dayOfMonth"] = flowdynamo.NInt(*req.DayOfMonth)
	}
	if req.AccountID != "" {
		item["accountId"] = flowdynamo.S(req.AccountID)
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{TableName: &s.tableName, Item: item}); err != nil {
		return "", err
	}
	return flowdynamo.SKSalary, nil
}

// ---------- reads ----------

// ListBudgets returns every BUDGET# item under the user's PLAN partition.
func (s *PlanningService) ListBudgets(ctx context.Context, userID string) ([]dto.BudgetResponse, error) {
	out, err := s.queryByPrefix(ctx, userID, flowdynamo.SKBudgetPrefix)
	if err != nil {
		return nil, err
	}
	budgets := make([]dto.BudgetResponse, 0, len(out))
	for _, item := range out {
		limit, _ := decimal.NewFromString(flowdynamo.Num(item, "limitValue"))
		budgets = append(budgets, dto.BudgetResponse{
			ID:         flowdynamo.Str(item, "id"),
			Category:   flowdynamo.Str(item, "category"),
			LimitType:  defaultStr(flowdynamo.Str(item, "limitType"), "ABSOLUTE"),
			LimitValue: limit,
		})
	}
	return budgets, nil
}

// ListGoals returns every GOAL# item.
func (s *PlanningService) ListGoals(ctx context.Context, userID string) ([]dto.GoalResponse, error) {
	out, err := s.queryByPrefix(ctx, userID, flowdynamo.SKGoalPrefix)
	if err != nil {
		return nil, err
	}
	goals := make([]dto.GoalResponse, 0, len(out))
	for _, item := range out {
		g := dto.GoalResponse{
			ID:   flowdynamo.Str(item, "id"),
			Name: flowdynamo.Str(item, "name"),
		}
		if v := flowdynamo.Num(item, "expectedReturnRate"); v != "" {
			d, _ := decimal.NewFromString(v)
			g.ExpectedReturnRate = &d
		}
		if v := flowdynamo.Num(item, "monthlyContribution"); v != "" {
			d, _ := decimal.NewFromString(v)
			g.MonthlyContribution = &d
		}
		if v := flowdynamo.Num(item, "targetAmount"); v != "" {
			d, _ := decimal.NewFromString(v)
			g.TargetAmount = &d
		}
		goals = append(goals, g)
	}
	return goals, nil
}

// GetEconomicParameters reads the single ECON#PARAMS item or returns zeros.
func (s *PlanningService) GetEconomicParameters(ctx context.Context, userID string) (dto.EconomicParametersResponse, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.UserPlanPK(userID)),
			"SK": flowdynamo.S(flowdynamo.SKEconParams),
		},
	})
	if err != nil {
		return dto.EconomicParametersResponse{}, err
	}
	if len(out.Item) == 0 {
		return dto.EconomicParametersResponse{SelicRate: decimal.Zero, IpcaRate: decimal.Zero}, nil
	}
	selic, _ := decimal.NewFromString(flowdynamo.Num(out.Item, "selicRate"))
	ipca, _ := decimal.NewFromString(flowdynamo.Num(out.Item, "ipcaRate"))
	return dto.EconomicParametersResponse{SelicRate: selic, IpcaRate: ipca}, nil
}

// GetSalary returns nil when no salary has been configured.
func (s *PlanningService) GetSalary(ctx context.Context, userID string) (*dto.SalaryResponse, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.UserPlanPK(userID)),
			"SK": flowdynamo.S(flowdynamo.SKSalary),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	amount, _ := decimal.NewFromString(flowdynamo.Num(out.Item, "amount"))
	resp := &dto.SalaryResponse{
		Amount:    amount,
		AccountID: flowdynamo.Str(out.Item, "accountId"),
	}
	if v := flowdynamo.Num(out.Item, "dayOfMonth"); v != "" {
		var i int
		_, _ = fmt.Sscanf(v, "%d", &i)
		resp.DayOfMonth = &i
	}
	return resp, nil
}

// queryByPrefix returns every item under the user's PLAN PK whose SK starts
// with `prefix` (used for budgets and goals).
func (s *PlanningService) queryByPrefix(ctx context.Context, userID, prefix string) ([]map[string]types.AttributeValue, error) {
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.tableName,
		KeyConditionExpression: strPtr("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": flowdynamo.S(flowdynamo.UserPlanPK(userID)),
			":sk": flowdynamo.S(prefix),
		},
	})
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
