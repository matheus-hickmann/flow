package service

import (
	"context"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

// RecoveryService stores the user's 3 recovery questions at
// PK=USER#{userId}, SK=RECOVERY.
type RecoveryService struct {
	dynamo    flowdynamo.API
	tableName string
}

// NewRecoveryService wires the service.
func NewRecoveryService(dynamo flowdynamo.API, tableName string) *RecoveryService {
	return &RecoveryService{dynamo: dynamo, tableName: tableName}
}

const skRecovery = "RECOVERY"

// Save replaces the user's recovery questions (full overwrite).
func (s *RecoveryService) Save(ctx context.Context, userID string, items []dto.RecoveryQuestionItem) error {
	list := make([]types.AttributeValue, 0, len(items))
	for _, q := range items {
		list = append(list, &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"question": flowdynamo.S(q.Question),
			"answer":   flowdynamo.S(q.Answer),
		}})
	}
	_, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.tableName,
		Item: map[string]types.AttributeValue{
			"PK":        flowdynamo.S(flowdynamo.UserPK(userID)),
			"SK":        flowdynamo.S(skRecovery),
			"type":      flowdynamo.S("RECOVERY"),
			"items":     &types.AttributeValueMemberL{Value: list},
			"updatedAt": flowdynamo.S(time.Now().UTC().Format(time.RFC3339Nano)),
		},
	})
	return err
}

// GetQuestionsOnly returns the question texts (not the answers).
func (s *RecoveryService) GetQuestionsOnly(ctx context.Context, userID string) ([]string, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.UserPK(userID)),
			"SK": flowdynamo.S(skRecovery),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return []string{}, nil
	}
	l, ok := out.Item["items"].(*types.AttributeValueMemberL)
	if !ok {
		return []string{}, nil
	}
	questions := make([]string, 0, len(l.Value))
	for _, raw := range l.Value {
		m, ok := raw.(*types.AttributeValueMemberM)
		if !ok {
			continue
		}
		if q, ok := m.Value["question"].(*types.AttributeValueMemberS); ok {
			questions = append(questions, q.Value)
		}
	}
	return questions, nil
}
