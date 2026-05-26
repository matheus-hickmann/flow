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

var ErrGroupNotFound = errors.New("group not found")
var ErrNotGroupOwner = errors.New("only the group owner can perform this action")
var ErrAlreadyMember = errors.New("user is already a member of this group")

// GroupService manages family groups and their memberships.
type GroupService struct {
	dynamo   flowdynamo.API
	table    string
	accounts *AccountService
}

// NewGroupService wires the service.
func NewGroupService(dynamo flowdynamo.API, table string, accounts *AccountService) *GroupService {
	return &GroupService{dynamo: dynamo, table: table, accounts: accounts}
}

// Create creates a new group owned by the caller and adds them as the first member.
func (s *GroupService) Create(ctx context.Context, ownerID, ownerName string, req dto.CreateGroupRequest) (dto.GroupResponse, error) {
	groupID := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	gPK := flowdynamo.GroupPK(groupID)

	// Group metadata item
	groupItem := map[string]types.AttributeValue{
		"PK":        flowdynamo.S(gPK),
		"SK":        flowdynamo.S(flowdynamo.SKMetadata),
		"type":      flowdynamo.S("GROUP"),
		"id":        flowdynamo.S(groupID),
		"name":      flowdynamo.S(req.Name),
		"ownerId":   flowdynamo.S(ownerID),
		"createdAt": flowdynamo.S(now),
		"GSI1PK":    flowdynamo.S(flowdynamo.GSI1PKUserGroups(ownerID)),
		"GSI1SK":    flowdynamo.S(flowdynamo.GSI1SKUserGroup(groupID)),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.table, Item: groupItem,
	}); err != nil {
		return dto.GroupResponse{}, err
	}

	// Owner membership item (inside group PK for efficient member listing)
	memberItem := map[string]types.AttributeValue{
		"PK":          flowdynamo.S(gPK),
		"SK":          flowdynamo.S(flowdynamo.MemberSK(ownerID)),
		"type":        flowdynamo.S("GROUP_MEMBER"),
		"userId":      flowdynamo.S(ownerID),
		"displayName": flowdynamo.S(ownerName),
		"role":        flowdynamo.S("owner"),
		"joinedAt":    flowdynamo.S(now),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.table, Item: memberItem,
	}); err != nil {
		return dto.GroupResponse{}, err
	}

	// User → group reverse index (for "list my groups")
	userGroupItem := map[string]types.AttributeValue{
		"PK":        flowdynamo.S(flowdynamo.UserPK(ownerID)),
		"SK":        flowdynamo.S(flowdynamo.GSI1SKUserGroup(groupID)),
		"type":      flowdynamo.S("USER_GROUP"),
		"groupId":   flowdynamo.S(groupID),
		"groupName": flowdynamo.S(req.Name),
		"role":      flowdynamo.S("owner"),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.table, Item: userGroupItem,
	}); err != nil {
		return dto.GroupResponse{}, err
	}

	return dto.GroupResponse{
		ID:      groupID,
		Name:    req.Name,
		OwnerID: ownerID,
		Members: []dto.MemberResponse{{UserID: ownerID, DisplayName: ownerName, Role: "owner", JoinedAt: now}},
		CreatedAt: now,
	}, nil
}

// ListForUser returns all groups the user belongs to.
func (s *GroupService) ListForUser(ctx context.Context, userID string) ([]dto.GroupResponse, error) {
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.table,
		KeyConditionExpression: strPtr("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     flowdynamo.S(flowdynamo.UserPK(userID)),
			":prefix": flowdynamo.S("GROUP#"),
		},
	})
	if err != nil {
		return nil, err
	}
	groups := make([]dto.GroupResponse, 0, len(out.Items))
	for _, item := range out.Items {
		groups = append(groups, dto.GroupResponse{
			ID:   flowdynamo.Str(item, "groupId"),
			Name: flowdynamo.Str(item, "groupName"),
		})
	}
	return groups, nil
}

// Get returns a single group with its full member list.
func (s *GroupService) Get(ctx context.Context, groupID string) (*dto.GroupResponse, error) {
	// Fetch metadata
	meta, err := s.dynamo.GetItem(ctx, &awsdynamodb.GetItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.GroupPK(groupID)),
			"SK": flowdynamo.S(flowdynamo.SKMetadata),
		},
	})
	if err != nil {
		return nil, err
	}
	if len(meta.Item) == 0 {
		return nil, nil
	}

	members, err := s.listMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &dto.GroupResponse{
		ID:        flowdynamo.Str(meta.Item, "id"),
		Name:      flowdynamo.Str(meta.Item, "name"),
		OwnerID:   flowdynamo.Str(meta.Item, "ownerId"),
		CreatedAt: flowdynamo.Str(meta.Item, "createdAt"),
		Members:   members,
	}, nil
}

// AddMember writes the membership items for a newly accepted invite.
func (s *GroupService) AddMember(ctx context.Context, groupID, userID, displayName string) error {
	group, err := s.Get(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	for _, m := range group.Members {
		if m.UserID == userID {
			return ErrAlreadyMember
		}
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	gPK := flowdynamo.GroupPK(groupID)

	memberItem := map[string]types.AttributeValue{
		"PK":          flowdynamo.S(gPK),
		"SK":          flowdynamo.S(flowdynamo.MemberSK(userID)),
		"type":        flowdynamo.S("GROUP_MEMBER"),
		"userId":      flowdynamo.S(userID),
		"displayName": flowdynamo.S(displayName),
		"role":        flowdynamo.S("member"),
		"joinedAt":    flowdynamo.S(now),
	}
	if _, err := s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.table, Item: memberItem,
	}); err != nil {
		return err
	}

	userGroupItem := map[string]types.AttributeValue{
		"PK":        flowdynamo.S(flowdynamo.UserPK(userID)),
		"SK":        flowdynamo.S(flowdynamo.GSI1SKUserGroup(groupID)),
		"type":      flowdynamo.S("USER_GROUP"),
		"groupId":   flowdynamo.S(groupID),
		"groupName": flowdynamo.S(group.Name),
		"role":      flowdynamo.S("member"),
	}
	_, err = s.dynamo.PutItem(ctx, &awsdynamodb.PutItemInput{
		TableName: &s.table, Item: userGroupItem,
	})
	return err
}

// RemoveMember deletes the membership items for a given user.
func (s *GroupService) RemoveMember(ctx context.Context, groupID, callerID, targetUserID string) error {
	group, err := s.Get(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	if callerID != targetUserID && group.OwnerID != callerID {
		return ErrNotGroupOwner
	}

	// Delete from group PK
	if _, err := s.dynamo.DeleteItem(ctx, &awsdynamodb.DeleteItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.GroupPK(groupID)),
			"SK": flowdynamo.S(flowdynamo.MemberSK(targetUserID)),
		},
	}); err != nil {
		return err
	}

	// Delete reverse index
	_, err = s.dynamo.DeleteItem(ctx, &awsdynamodb.DeleteItemInput{
		TableName: &s.table,
		Key: map[string]types.AttributeValue{
			"PK": flowdynamo.S(flowdynamo.UserPK(targetUserID)),
			"SK": flowdynamo.S(flowdynamo.GSI1SKUserGroup(groupID)),
		},
	})
	return err
}

// Delete removes the group and all its member/invite items (owner only).
func (s *GroupService) Delete(ctx context.Context, groupID, callerID string) error {
	group, err := s.Get(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	if group.OwnerID != callerID {
		return ErrNotGroupOwner
	}

	// Remove all member reverse-index entries first
	for _, m := range group.Members {
		_ = s.RemoveMember(ctx, groupID, callerID, m.UserID)
	}

	// Scan and delete everything under GROUP#<gid> PK
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.table,
		KeyConditionExpression: strPtr("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": flowdynamo.S(flowdynamo.GroupPK(groupID)),
		},
		ProjectionExpression: strPtr("SK"),
	})
	if err != nil {
		return err
	}
	for _, item := range out.Items {
		sk := flowdynamo.Str(item, "SK")
		if _, err := s.dynamo.DeleteItem(ctx, &awsdynamodb.DeleteItemInput{
			TableName: &s.table,
			Key: map[string]types.AttributeValue{
				"PK": flowdynamo.S(flowdynamo.GroupPK(groupID)),
				"SK": flowdynamo.S(sk),
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

// ListSharedAccounts returns all shared=true accounts from every member of the group.
func (s *GroupService) ListSharedAccounts(ctx context.Context, callerID, groupID string) ([]dto.SharedAccountResponse, error) {
	group, err := s.Get(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}
	// Verify caller is a member
	isMember := false
	for _, m := range group.Members {
		if m.UserID == callerID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, ErrNotGroupOwner
	}

	var result []dto.SharedAccountResponse
	for _, member := range group.Members {
		if member.UserID == callerID {
			continue // skip own accounts in this view
		}
		accounts, err := s.accounts.ListFiltered(ctx, member.UserID, false)
		if err != nil {
			continue
		}
		for _, acc := range accounts {
			if acc.Shared {
				result = append(result, dto.SharedAccountResponse{
					AccountResponse: acc,
					OwnerID:         member.UserID,
					OwnerName:       member.DisplayName,
				})
			}
		}
	}
	return result, nil
}

// ---------- helpers ----------

func (s *GroupService) listMembers(ctx context.Context, groupID string) ([]dto.MemberResponse, error) {
	out, err := s.dynamo.Query(ctx, &awsdynamodb.QueryInput{
		TableName:              &s.table,
		KeyConditionExpression: strPtr("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     flowdynamo.S(flowdynamo.GroupPK(groupID)),
			":prefix": flowdynamo.S("MEMBER#"),
		},
	})
	if err != nil {
		return nil, err
	}
	members := make([]dto.MemberResponse, 0, len(out.Items))
	for _, item := range out.Items {
		members = append(members, dto.MemberResponse{
			UserID:      flowdynamo.Str(item, "userId"),
			DisplayName: flowdynamo.Str(item, "displayName"),
			Role:        flowdynamo.Str(item, "role"),
			JoinedAt:    flowdynamo.Str(item, "joinedAt"),
		})
	}
	return members, nil
}
