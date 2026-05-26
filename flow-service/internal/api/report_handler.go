package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/service"
)

type reportHandler struct {
	reports *service.ReportService
}

func newReportHandler(s *service.ReportService) *reportHandler {
	return &reportHandler{reports: s}
}

func (h *reportHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)
	r.Get("/monthly", h.monthly)
	return r
}

func (h *reportHandler) monthly(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	reportType := q.Get("type")
	if reportType == "" {
		reportType = "expense"
	}
	got, err := h.reports.Monthly(
		r.Context(),
		middleware.UserIDFrom(r.Context()),
		q.Get("from"),
		q.Get("to"),
		reportType,
		q.Get("expenseAccountId"),
		q.Get("incomeAccountId"),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, got)
}
