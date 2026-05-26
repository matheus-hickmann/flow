package service

import (
	"context"
	"errors"
	"time"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/hickmann/flow-service/internal/dto"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
)

var ErrInviteNotFound = errors.New("invite not found")
var ErrInviteExpired = errors.New("invite has expired")
var ErrInviteRevoked = errors.New("invite has been revoked")
var ErrInviteAlreadyUsed = errors.New("invite has already been used")

const inviteTTL = 24 * time.Hour

// InviteService handles token-based group invites.
type InviteService struct {
	dynamo  flowdynamo.API
	table   string
	groups  *GroupService
}

// NewInviteService wires the service.
func NewInviteService(dynamo flowdynamo.API, table string, groups *GroupService) *InviteService {
	return &InviteService{dynamo: dynamo, table: table, groups: groups}
}

// GenerateInvite creates a new pending invite for a group (owner only).
func (s *InviteService) GenerateInvite(ctx context.Context, groupID, callerID, inviterName, inviteeLabel string) (dto.InviteResponse, error) {
	group, err := s.groups.Get(ctx, groupID)
	if err != nil {
		return dto.InviteResponse{}, err
	}
	if group == nil {
		return dto.InviteResponse{}, ErrGroupNotFound
	}
	if group.OwnerID != callerID {
		return dto.InviteResponse{}, ErrNotGroupOwner
	}

	token := uuid.NewString()
	now := time.Now().UTC()
	expiresAt := now.Add(inviteTTL).Format(time.RFC3339Nano)
	nowStr := now.Format(time.RFC3339Nano)

	inviteItem := map[string]types.AttributeValue{
		"PK":           flowdynamo.S(flowdynamo.InvitePK(token)),
		"SK":           flowdynamo.S(flowdynamo.SKMetadata),
		"type":         flowdynamo.S("INVITE"),
		"token":        flowdynamo.S(token),
		"groupId":      flowdynamo.S(groupID),
		"groupName":    flowdynamo.S(group.Name),
		"inviteeLabel": flowdynamo.S(inviteeLabel),
		"inviterName":  flowdynamo.S(inviterName),
		"status":       flowdynamo.S("pending"),
		"createdAt":    flowdynamo.S(nowStr),
		"expiresAt":    flowdynamo.S(expiresAt),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.table, Item: inviteItem,
	}); err != nil {
		return dto.InviteResponse{}, err
	}

	// Reference item inside the group PK for listing
	groupInviteItem := map[string]types.AttributeValue{
		"PK":           flowdynamo.S(flowdynamo.GroupPK(groupID)),
		"SK":           flowdynamo.S(flowdynamo.InviteSK(token)),
		"type":         flowdynamo.S("GROUP_INVITE"),
		"token":        flowdynamo.S(token),
		"inviteeLabel": flowdynamo.S(inviteeLabel),
		"inviterName":  flowdynamo.S(inviterName),
		"status":       flowdynamo.S("pending"),
		"createdAt":    flowdynamo.S(nowStr),
		"expiresAt":    flowdynamo.S(expiresAt),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.table, Item: groupInviteItem,
	}); err != nil {
		return dto.InviteResponse{}, err
	}

	return dto.InviteResponse{
		Token:        token,
		GroupID:      groupID,
		GroupName:    group.Name,
		InviteeLabel: inviteeLabel,
		InviterName:  inviterName,
		ExpiresAt:    expiresAt,
		Status:       "pending",
	}, nil
}

// GetPreview returns public invite details without authentication.
func (s *InviteService) GetPreview(ctx context.Context, token string) (dto.InvitePreviewResponse, error) {
	item, err := s.fetchInviteItem(ctx, token)
	if err != nil {
		return dto.InvitePreviewResponse{}, err
	}
	if item == nil {
		return dto.InvitePreviewResponse{Valid: false}, nil
	}

	status := flowdynamo.Str(item, "status")
	expiresAt := flowdynamo.Str(item, "expiresAt")
	expired := isExpired(expiresAt)

	valid := status == "pending" && !expired
	return dto.InvitePreviewResponse{
		GroupName:   flowdynamo.Str(item, "groupName"),
		InviterName: flowdynamo.Str(item, "inviterName"),
		ExpiresAt:   expiresAt,
		Valid:       valid,
	}, nil
}

// Accept validates the invite and adds the caller as a group member.
func (s *InviteService) Accept(ctx context.Context, token, userID, displayName string) error {
	item, err := s.fetchInviteItem(ctx, token)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrInviteNotFound
	}

	switch flowdynamo.Str(item, "status") {
	case "revoked":
		return ErrInviteRevoked
	case "accepted":
		return ErrInviteAlreadyUsed
	}
	if isExpired(flowdynamo.Str(item, "expiresAt")) {
		return ErrInviteExpired
	}

	groupID := flowdynamo.Str(item, "groupId")
	if err := s.groups.AddMember(ctx, groupID, userID, displayName); err != nil {
		return err
	}

	return s.updateStatus(ctx, token, groupID, "accepted")
}

// Revoke cancels a pending invite (owner only).
func (s *InviteService) Revoke(ctx context.Context, token, callerID string) error {
	item, err := s.fetchInviteItem(ctx, token)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrInviteNotFound
	}

	groupID := flowdynamo.Str(item, "groupId")
	group, err := s.groups.Get(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	if group.OwnerID != callerID {
		return ErrNotGroupOwner
	}

	return s.updateStatus(ctx, token, groupID, "revoked")
}

// ListForGroup returns all invites for a group (owner only).
func (s *InviteService) ListForGroup(ctx context.Context, groupID, callerID string) ([]dto.InviteResponse, error) {
	group, err := s.groups.Get(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}
	if group.OwnerID != callerID {
		return nil, ErrNotGroupOwner
	}

	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.table,
		KeyConditionExpression: strPtr("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     flowdynamo.S(flowdynamo.GroupPK(groupID)),
			":prefix": flowdynamo.S("INVITE#"),
		},
	})
	if err != nil {
		return nil, err
	}

	invites := make([]dto.InviteResponse, 0, len(out.Items))
	for _, i := range out.Items {
		invites = append(invites, dto.InviteResponse{
			Token:        flowdynamo.Str(i, "token"),
			GroupID:      groupID,
			GroupName:    group.Name,
			InviteeLabel: flowdynamo.Str(i, "inviteeLabel"),
			InviterName:  flowdynamo.Str(i, "inviterName"),
			ExpiresAt:    flowdynamo.Str(i, "expiresAt"),
			Status:       flowdynamo.Str(i, "status"),
		})
	}
	return invites, nil
}

// ---------- helpers ----------

func (s *InviteService) fetchInviteItem(ctx context.Context, token string) (map[string]types.AttributeValue, error) {
	out, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.InvitePK(token)),
			"SK": flowdynamo.S(flowdynamo.SKMetadata),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Item) == 0 {
		return nil, nil
	}
	return out.Item, nil
}

func (s *InviteService) updateStatus(ctx context.Context, token, groupID, status string) error {
	updateExpr := strPtr("SET #s = :status")
	attrNames := map[string]string{"#s": "status"}
	attrValues := map[string]types.AttributeValue{":status": flowdynamo.S(status)}

	// Update canonical item
	if _, err := s.dynamo.UpdateItem(ctx, &awsdynamodb.UpdateItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.InvitePK(token)),
			"SK": flowdynamo.S(flowdynamo.SKMetadata),
		},
		UpdateExpression:          updateExpr,
		ExpressionAttributeNames:  attrNames,
		ExpressionAttributeValues: attrValues,
	}); err != nil {
		return err
	}

	// Update reference item inside the group
	_, err := s.dynamo.UpdateItem(ctx, &awsdynamodb.UpdateItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.GroupPK(groupID)),
			"SK": flowdynamo.S(flowdynamo.InviteSK(token)),
		},
		UpdateExpression:          updateExpr,
		ExpressionAttributeNames:  attrNames,
		ExpressionAttributeValues: attrValues,
	})
	return err
}

func isExpired(expiresAt string) bool {
	t, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return true
	}
	return time.Now().UTC().After(t)
}
