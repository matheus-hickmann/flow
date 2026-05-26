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

type planningHandler struct {
	planning *service.PlanningService
}

func newPlanningHandler(s *service.PlanningService) *planningHandler {
	return &planningHandler{planning: s}
}

func (h *planningHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)
	r.Post("/submit", h.submit)
	r.Get("/budgets", h.listBudgets)
	r.Get("/goals", h.listGoals)
	r.Get("/economic-parameters", h.getEconomicParameters)
	r.Get("/salary", h.getSalary)
	return r
}

func (h *planningHandler) submit(w http.ResponseWriter, r *http.Request) {
	var req dto.PlanningSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	id, err := h.planning.Submit(r.Context(), middleware.UserIDFrom(r.Context()), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPlanning) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *planningHandler) listBudgets(w http.ResponseWriter, r *http.Request) {
	budgets, err := h.planning.ListBudgets(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, budgets)
}

func (h *planningHandler) listGoals(w http.ResponseWriter, r *http.Request) {
	goals, err := h.planning.ListGoals(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, goals)
}

func (h *planningHandler) getEconomicParameters(w http.ResponseWriter, r *http.Request) {
	got, err := h.planning.GetEconomicParameters(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}

func (h *planningHandler) getSalary(w http.ResponseWriter, r *http.Request) {
	salary, err := h.planning.GetSalary(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if salary == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusOK, salary)
}
