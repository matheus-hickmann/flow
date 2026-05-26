package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/dto"
	"github.com/hickmann/flow-service/internal/service"
)

// authHandler bundles the auth-related HTTP routes.
type authHandler struct {
	auth    *service.AuthService
	devMode bool
}

func newAuthHandler(s *service.AuthService, devMode bool) *authHandler {
	return &authHandler{auth: s, devMode: devMode}
}

func (h *authHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/signup", h.signup)
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)
	return r
}

func (h *authHandler) signup(w http.ResponseWriter, r *http.Request) {
	var req dto.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.UserID == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "userId and password are required")
		return
	}

	user, err := h.auth.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := middleware.IssueToken(user.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token generation failed")
		return
	}
	setAuthCookie(w, token, h.devMode)
	writeJSON(w, http.StatusCreated, dto.AuthResponse{
		AccessToken: token,
		UserID:      user.UserID,
		DisplayName: user.DisplayName,
	})
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.UserID == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "userId and password are required")
		return
	}

	user, err := h.auth.ValidateLogin(r.Context(), req.UserID, req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "ID de usuário ou senha inválidos")
		return
	}

	token, err := middleware.IssueToken(user.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token generation failed")
		return
	}
	setAuthCookie(w, token, h.devMode)
	writeJSON(w, http.StatusOK, dto.AuthResponse{
		AccessToken: token,
		UserID:      user.UserID,
		DisplayName: user.DisplayName,
	})
}

func (h *authHandler) logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   !h.devMode,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// setAuthCookie writes the HttpOnly access_token cookie.
// Secure is disabled in dev mode (HTTP) and enabled in production (HTTPS).
func setAuthCookie(w http.ResponseWriter, token string, devMode bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
		Secure:   !devMode,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   86400, // 24 hours
	})
}

// writeError emits a JSON {"message": ...} body.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}
