package service

import (
	"context"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

// CategoryService persists the user's expense/income category lists at
// PK=USER#{userId}, SK=CATEGORIES. Returns defaults when not yet customized.
type CategoryService struct {
	dynamo    flowdynamo.API
	tableName string
}

// NewCategoryService wires the service.
func NewCategoryService(dynamo flowdynamo.API, tableName string) *CategoryService {
	return &CategoryService{dynamo: dynamo, tableName: tableName}
}

// Get returns the user's category lists or DefaultCategoryList when absent.
func (s *CategoryService) Get(ctx context.Context, userID string) (dto.CategoryList, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.UserPK(userID)),
			"SK": flowdynamo.S(flowdynamo.SKCategories),
		},
	})
	if err != nil {
		return dto.CategoryList{}, err
	}
	if len(out.Item) == 0 {
		return dto.DefaultCategoryList(), nil
	}
	return dto.CategoryList{
		Expense: parseCategoryList(out.Item["expense"]),
		Income:  parseCategoryList(out.Item["income"]),
	}, nil
}

// Save replaces the user's category lists entirely (idempotent overwrite).
func (s *CategoryService) Save(ctx context.Context, userID string, payload dto.CategoryList) (dto.CategoryList, error) {
	_, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.tableName,
		Item: map[string]types.AttributeValue{
			"PK":        flowdynamo.S(flowdynamo.UserPK(userID)),
			"SK":        flowdynamo.S(flowdynamo.SKCategories),
			"type":      flowdynamo.S("CATEGORIES"),
			"expense":   categoryListAttr(payload.Expense),
			"income":    categoryListAttr(payload.Income),
			"updatedAt": flowdynamo.S(time.Now().UTC().Format(time.RFC3339Nano)),
		},
	})
	if err != nil {
		return dto.CategoryList{}, err
	}
	return payload, nil
}

func categoryListAttr(items []dto.CategoryItem) types.AttributeValue {
	list := make([]types.AttributeValue, 0, len(items))
	for _, c := range items {
		list = append(list, &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"id":    flowdynamo.S(c.ID),
			"name":  flowdynamo.S(c.Name),
			"color": flowdynamo.S(c.Color),
		}})
	}
	return &types.AttributeValueMemberL{Value: list}
}

func parseCategoryList(av types.AttributeValue) []dto.CategoryItem {
	l, ok := av.(*types.AttributeValueMemberL)
	if !ok {
		return []dto.CategoryItem{}
	}
	out := make([]dto.CategoryItem, 0, len(l.Value))
	for _, raw := range l.Value {
		m, ok := raw.(*types.AttributeValueMemberM)
		if !ok {
			continue
		}
		out = append(out, dto.CategoryItem{
			ID:    flowdynamo.Str(m.Value, "id"),
			Name:  flowdynamo.Str(m.Value, "name"),
			Color: flowdynamo.Str(m.Value, "color"),
		})
	}
	return out
}
