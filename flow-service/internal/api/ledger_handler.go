package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/dto"
	"github.com/hickmann/flow-service/internal/service"
)

// ledgerHandler exposes /api/v1/ledger/* — accounts CRUD + transactions.
type ledgerHandler struct {
	accounts     *service.AccountService
	transactions *service.TransactionService
}

func newLedgerHandler(accounts *service.AccountService, transactions *service.TransactionService) *ledgerHandler {
	return &ledgerHandler{accounts: accounts, transactions: transactions}
}

func (h *ledgerHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)
	r.Route("/accounts", func(r chi.Router) {
		r.Get("/", h.listAccounts)
		r.Post("/", h.createAccount)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.getAccount)
			r.Patch("/", h.updateAccount)
			r.Delete("/", h.deleteAccount)
			r.Get("/balance", h.getBalance)
			r.Patch("/balance", h.adjustBalance)
		})
	})
	r.Route("/transactions", func(r chi.Router) {
		r.Get("/", h.listTransactions)
		r.Post("/", h.postTransaction)
	})
	return r
}

// ---------- accounts ----------

func (h *ledgerHandler) listAccounts(w http.ResponseWriter, r *http.Request) {
	includeSystem := r.URL.Query().Get("includeSystem") == "true"
	accounts, err := h.accounts.ListFiltered(r.Context(), middleware.UserIDFrom(r.Context()), includeSystem)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (h *ledgerHandler) createAccount(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	acc, err := h.accounts.Create(r.Context(), middleware.UserIDFrom(r.Context()), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, acc)
}

func (h *ledgerHandler) getAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	acc, err := h.accounts.GetByID(r.Context(), middleware.UserIDFrom(r.Context()), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if acc == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (h *ledgerHandler) updateAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	acc, err := h.accounts.Update(r.Context(), middleware.UserIDFrom(r.Context()), id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAccountNotFound):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrSystemAccountReadOnly):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (h *ledgerHandler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.accounts.Delete(r.Context(), middleware.UserIDFrom(r.Context()), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ledgerHandler) getBalance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	balance, err := h.accounts.GetBalanceSnapshot(r.Context(), middleware.UserIDFrom(r.Context()), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if balance == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, balance)
}

func (h *ledgerHandler) adjustBalance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.AdjustBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	acc, err := h.accounts.AdjustBalance(r.Context(), middleware.UserIDFrom(r.Context()), id, req)
	if err != nil {
		if errors.Is(err, service.ErrAccountNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

// ---------- transactions ----------

func (h *ledgerHandler) listTransactions(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	accountID := r.URL.Query().Get("accountId")
	txns, err := h.transactions.List(r.Context(), middleware.UserIDFrom(r.Context()), limit, accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, txns)
}

func (h *ledgerHandler) postTransaction(w http.ResponseWriter, r *http.Request) {
	var req dto.PostTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	tx, err := h.transactions.Post(r.Context(), middleware.UserIDFrom(r.Context()), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTransaction) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}
