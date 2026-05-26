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

type debtHandler struct {
	debt *service.DebtService
}

func newDebtHandler(s *service.DebtService) *debtHandler {
	return &debtHandler{debt: s}
}

func (h *debtHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Post("/{id}/payment", h.recordPayment)
	r.Delete("/{id}", h.delete)
	return r
}

func (h *debtHandler) list(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	debts, err := h.debt.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, debts)
}

func (h *debtHandler) create(w http.ResponseWriter, r *http.Request) {
	var req dto.DebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	userID := middleware.UserIDFrom(r.Context())
	id, err := h.debt.Create(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidDebt) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *debtHandler) get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	debt, err := h.debt.Get(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, service.ErrDebtNotFound) {
			writeError(w, http.StatusNotFound, "debt not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, debt)
}

func (h *debtHandler) recordPayment(w http.ResponseWriter, r *http.Request) {
	var req dto.DebtPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	userID := middleware.UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	debt, err := h.debt.RecordPayment(r.Context(), userID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDebtNotFound):
			writeError(w, http.StatusNotFound, "debt not found")
		case errors.Is(err, service.ErrDebtOverpaid):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, service.ErrInvalidDebt):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, debt)
}

func (h *debtHandler) delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.debt.Delete(r.Context(), userID, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
