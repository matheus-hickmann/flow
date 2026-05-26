package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/service"
)

type inviteHandler struct {
	invites *service.InviteService
	auth    *service.AuthService
}

func newInviteHandler(invites *service.InviteService, auth *service.AuthService) *inviteHandler {
	return &inviteHandler{invites: invites, auth: auth}
}

func (h *inviteHandler) routes() http.Handler {
	r := chi.NewRouter()
	// Preview is public — no auth middleware on this sub-router
	r.Get("/{token}", h.preview)
	// Accept requires a valid token
	r.With(middleware.RequireAuth).Post("/{token}/accept", h.accept)
	return r
}

func (h *inviteHandler) preview(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	resp, err := h.invites.GetPreview(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *inviteHandler) accept(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	token := chi.URLParam(r, "token")

	var body struct {
		DisplayName string `json:"displayName"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	displayName := body.DisplayName
	if displayName == "" {
		if user, err := h.auth.FindByUserID(r.Context(), userID); err == nil && user != nil && user.DisplayName != "" {
			displayName = user.DisplayName
		} else {
			displayName = userID
		}
	}

	if err := h.invites.Accept(r.Context(), token, userID, displayName); err != nil {
		writeError(w, errStatus(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
