package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/service"
)

type dashboardHandler struct {
	dashboard *service.DashboardService
}

func newDashboardHandler(s *service.DashboardService) *dashboardHandler {
	return &dashboardHandler{dashboard: s}
}

func (h *dashboardHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)
	r.Get("/summary", h.summary)
	r.Get("/summary/monthly", h.monthlySummary)
	r.Get("/planned-vs-actual", h.plannedVsActual)
	r.Get("/balance-trend", h.balanceTrend)
	r.Get("/projection", h.projection)
	r.Get("/daily-expenses", h.dailyExpenses)
	r.Get("/budget-projection", h.budgetProjection)
	return r
}

func (h *dashboardHandler) dailyExpenses(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	got, err := h.dashboard.GetDailyExpenses(r.Context(), middleware.UserIDFrom(r.Context()), year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}

func (h *dashboardHandler) projection(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	got, err := h.dashboard.GetProjection(r.Context(), middleware.UserIDFrom(r.Context()), year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}

func (h *dashboardHandler) balanceTrend(w http.ResponseWriter, r *http.Request) {
	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			days = n
		}
	}
	points, err := h.dashboard.GetBalanceTrend(r.Context(), middleware.UserIDFrom(r.Context()), days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, points)
}

func (h *dashboardHandler) summary(w http.ResponseWriter, r *http.Request) {
	got, err := h.dashboard.GetSummary(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}

func (h *dashboardHandler) monthlySummary(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	got, err := h.dashboard.GetMonthlySummary(r.Context(), middleware.UserIDFrom(r.Context()), year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}

func (h *dashboardHandler) budgetProjection(w http.ResponseWriter, r *http.Request) {
	months := 6
	if v := r.URL.Query().Get("months"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			months = n
		}
	}
	got, err := h.dashboard.GetBudgetProjection(r.Context(), middleware.UserIDFrom(r.Context()), months)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}

func (h *dashboardHandler) plannedVsActual(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	got, err := h.dashboard.GetPlannedVsActual(r.Context(), middleware.UserIDFrom(r.Context()), year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}
