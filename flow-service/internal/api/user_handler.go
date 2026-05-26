package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/dto"
	"github.com/hickmann/flow-service/internal/service"
)

// userHandler bundles /users/* routes (all protected).
type userHandler struct {
	auth     *service.AuthService
	recovery *service.RecoveryService
}

func newUserHandler(auth *service.AuthService, recovery *service.RecoveryService) *userHandler {
	return &userHandler{auth: auth, recovery: recovery}
}

func (h *userHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)
	r.Get("/me", h.me)
	r.Post("/me/recovery-questions", h.saveRecoveryQuestions)
	r.Get("/me/recovery-questions", h.getRecoveryQuestions)
	return r
}

func (h *userHandler) me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())

	resp := dto.MeResponse{UserID: userID, Source: "backend"}
	if user, err := h.auth.FindByUserID(r.Context(), userID); err == nil && user != nil && user.DisplayName != "" {
		resp.DisplayName = user.DisplayName
		resp.Name = user.DisplayName
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *userHandler) saveRecoveryQuestions(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRecoveryQuestionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if len(req.Questions) != 3 {
		writeError(w, http.StatusBadRequest, "É obrigatório cadastrar exatamente 3 perguntas de recuperação")
		return
	}
	for _, q := range req.Questions {
		if q.Question == "" || q.Answer == "" {
			writeError(w, http.StatusBadRequest, "question and answer are required")
			return
		}
	}

	if err := h.recovery.Save(r.Context(), middleware.UserIDFrom(r.Context()), req.Questions); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *userHandler) getRecoveryQuestions(w http.ResponseWriter, r *http.Request) {
	questions, err := h.recovery.GetQuestionsOnly(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"questions": questions})
}
