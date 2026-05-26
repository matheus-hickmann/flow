package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/dto"
	"github.com/hickmann/flow-service/internal/service"
)

type categoryHandler struct {
	categories *service.CategoryService
}

func newCategoryHandler(s *service.CategoryService) *categoryHandler {
	return &categoryHandler{categories: s}
}

func (h *categoryHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)
	r.Get("/", h.get)
	r.Put("/", h.save)
	return r
}

func (h *categoryHandler) get(w http.ResponseWriter, r *http.Request) {
	got, err := h.categories.Get(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}

func (h *categoryHandler) save(w http.ResponseWriter, r *http.Request) {
	var payload dto.CategoryList
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if payload.Expense == nil || payload.Income == nil {
		writeError(w, http.StatusBadRequest, "expense and income are required")
		return
	}
	saved, err := h.categories.Save(r.Context(), middleware.UserIDFrom(r.Context()), payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, saved)
}
