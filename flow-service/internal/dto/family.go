package dto

// ── Groups ────────────────────────────────────────────────────────────────────

// CreateGroupRequest is the body of POST /api/v1/groups.
type CreateGroupRequest struct {
	Name string `json:"name"`
}

// GroupResponse is returned for group detail and list endpoints.
type GroupResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	OwnerID   string           `json:"ownerId"`
	CreatedAt string           `json:"createdAt"`
	Members   []MemberResponse `json:"members,omitempty"`
}

// MemberResponse describes one member inside a group.
type MemberResponse struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"` // "owner" | "member"
	JoinedAt    string `json:"joinedAt"`
}

// ── Invites ───────────────────────────────────────────────────────────────────

// CreateInviteRequest is the body of POST /api/v1/groups/{id}/invites.
type CreateInviteRequest struct {
	InviteeLabel string `json:"inviteeLabel"` // free-text, e.g. "Ana - esposa"
}

// InviteResponse is returned after invite creation.
type InviteResponse struct {
	Token        string `json:"token"`
	GroupID      string `json:"groupId"`
	GroupName    string `json:"groupName"`
	InviteeLabel string `json:"inviteeLabel"`
	InviterName  string `json:"inviterName"`
	ExpiresAt    string `json:"expiresAt"`
	Status       string `json:"status"` // "pending" | "accepted" | "revoked"
}

// InvitePreviewResponse is the public (unauthenticated) preview of an invite.
type InvitePreviewResponse struct {
	GroupName   string `json:"groupName"`
	InviterName string `json:"inviterName"`
	ExpiresAt   string `json:"expiresAt"`
	Valid        bool   `json:"valid"`
}

// ── Shared accounts ───────────────────────────────────────────────────────────

// SharedAccountResponse adds ownerName to a regular account response for the group view.
type SharedAccountResponse struct {
	AccountResponse
	OwnerID   string `json:"ownerId"`
	OwnerName string `json:"ownerName"`
}
