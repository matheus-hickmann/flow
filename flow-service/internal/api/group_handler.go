package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/dto"
	"github.com/hickmann/flow-service/internal/service"
)

type groupHandler struct {
	groups  *service.GroupService
	invites *service.InviteService
	auth    *service.AuthService
}

func newGroupHandler(groups *service.GroupService, invites *service.InviteService, auth *service.AuthService) *groupHandler {
	return &groupHandler{groups: groups, invites: invites, auth: auth}
}

func (h *groupHandler) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)

	// Groups
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{groupId}", h.get)
	r.Delete("/{groupId}", h.delete)

	// Members
	r.Delete("/{groupId}/members/{userId}", h.removeMember)

	// Shared accounts
	r.Get("/{groupId}/accounts", h.sharedAccounts)

	// Invites
	r.Post("/{groupId}/invites", h.createInvite)
	r.Get("/{groupId}/invites", h.listInvites)
	r.Delete("/{groupId}/invites/{token}", h.revokeInvite)

	return r
}

func (h *groupHandler) create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	var req dto.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	displayName := h.resolveDisplayName(r, userID)
	resp, err := h.groups.Create(r.Context(), userID, displayName, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *groupHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	groups, err := h.groups.ListForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

func (h *groupHandler) get(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupId")
	group, err := h.groups.Get(r.Context(), groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if group == nil {
		writeError(w, http.StatusNotFound, "group not found")
		return
	}
	writeJSON(w, http.StatusOK, group)
}

func (h *groupHandler) delete(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFrom(r.Context())
	groupID := chi.URLParam(r, "groupId")
	if err := h.groups.Delete(r.Context(), groupID, callerID); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *groupHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFrom(r.Context())
	groupID := chi.URLParam(r, "groupId")
	targetUserID := chi.URLParam(r, "userId")
	if err := h.groups.RemoveMember(r.Context(), groupID, callerID, targetUserID); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *groupHandler) sharedAccounts(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFrom(r.Context())
	groupID := chi.URLParam(r, "groupId")
	accounts, err := h.groups.ListSharedAccounts(r.Context(), callerID, groupID)
	if err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h *groupHandler) createInvite(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFrom(r.Context())
	groupID := chi.URLParam(r, "groupId")
	var req dto.CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	inviterName := h.resolveDisplayName(r, callerID)
	resp, err := h.invites.GenerateInvite(r.Context(), groupID, callerID, inviterName, req.InviteeLabel)
	if err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *groupHandler) listInvites(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFrom(r.Context())
	groupID := chi.URLParam(r, "groupId")
	invites, err := h.invites.ListForGroup(r.Context(), groupID, callerID)
	if err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, invites)
}

func (h *groupHandler) revokeInvite(w http.ResponseWriter, r *http.Request) {
	callerID := middleware.UserIDFrom(r.Context())
	token := chi.URLParam(r, "token")
	if err := h.invites.Revoke(r.Context(), token, callerID); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// resolveDisplayName fetches the user's display name from the auth service,
// falling back to the userID if not found.
func (h *groupHandler) resolveDisplayName(r *http.Request, userID string) string {
	if user, err := h.auth.FindByUserID(r.Context(), userID); err == nil && user != nil && user.DisplayName != "" {
		return user.DisplayName
	}
	return userID
}

// errStatus maps domain errors to HTTP status codes.
func errStatus(err error) int {
	switch err {
	case service.ErrGroupNotFound, service.ErrInviteNotFound:
		return http.StatusNotFound
	case service.ErrNotGroupOwner:
		return http.StatusForbidden
	case service.ErrAlreadyMember, service.ErrInviteAlreadyUsed:
		return http.StatusConflict
	case service.ErrInviteExpired, service.ErrInviteRevoked:
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}
