// Package service contains the business logic shared by handlers.
package service

import (
	"context"
	"errors"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

// ErrUserAlreadyExists is returned by Create when the userId is taken.
var ErrUserAlreadyExists = errors.New("ID de usuário já em uso")

// AuthUser is the in-memory shape returned by AuthService lookups.
type AuthUser struct {
	UserID      string
	DisplayName string
}

// AuthService persists credentials under PK=AUTH#{userId}, SK=PROFILE.
// Passwords are stored as-is — for personal dev use; swap for Cognito in prod.
type AuthService struct {
	dynamo    flowdynamo.API
	tableName string
}

// NewAuthService wires the service.
func NewAuthService(dynamo flowdynamo.API, tableName string) *AuthService {
	return &AuthService{dynamo: dynamo, tableName: tableName}
}

// Create persists a new user. Returns ErrUserAlreadyExists if the id is taken.
func (s *AuthService) Create(ctx context.Context, req dto.SignupRequest) (AuthUser, error) {
	existing, err := s.FindByUserID(ctx, req.UserID)
	if err != nil {
		return AuthUser{}, err
	}
	if existing != nil {
		return AuthUser{}, ErrUserAlreadyExists
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	attrs := map[string]flowAV{
		"PK":        {S: flowdynamo.AuthPK(req.UserID)},
		"SK":        {S: flowdynamo.SKProfile},
		"type":      {S: "AUTH_USER"},
		"userId":    {S: req.UserID},
		"password":  {S: req.Password},
		"createdAt": {S: now},
		"updatedAt": {S: now},
	}
	if req.DisplayName != "" {
		attrs["displayName"] = flowAV{S: req.DisplayName}
	}

	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      buildItem(attrs),
	}); err != nil {
		return AuthUser{}, err
	}
	return AuthUser{UserID: req.UserID, DisplayName: req.DisplayName}, nil
}

// ValidateLogin returns the user when password matches; nil otherwise.
func (s *AuthService) ValidateLogin(ctx context.Context, userID, password string) (*AuthUser, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: buildItem(map[string]flowAV{
			"PK": {S: flowdynamo.AuthPK(userID)},
			"SK": {S: flowdynamo.SKProfile},
		}),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	if flowdynamo.Str(out.Item, "password") != password {
		return nil, nil
	}
	return &AuthUser{
		UserID:      flowdynamo.Str(out.Item, "userId"),
		DisplayName: flowdynamo.Str(out.Item, "displayName"),
	}, nil
}

// FindByUserID returns the user when present; nil + nil when not found.
func (s *AuthService) FindByUserID(ctx context.Context, userID string) (*AuthUser, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: buildItem(map[string]flowAV{
			"PK": {S: flowdynamo.AuthPK(userID)},
			"SK": {S: flowdynamo.SKProfile},
		}),
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	return &AuthUser{
		UserID:      flowdynamo.Str(out.Item, "userId"),
		DisplayName: flowdynamo.Str(out.Item, "displayName"),
	}, nil
}
