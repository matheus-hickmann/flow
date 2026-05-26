package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/dto"
	"github.com/hickmann/flow-service/internal/service"
)

const maxUploadBytes = 5 << 20 // 5 MB

type importHandler struct {
	imports *service.ImportService
}

func newImportHandler(s *service.ImportService) *importHandler {
	return &importHandler{imports: s}
}

func (h *importHandler) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequireAuth)
	r.Post("/parse", h.parse)
	r.Get("/merchant-rules", h.listRules)
	r.Post("/commit", h.commit)
	return r
}

// parse accepts a multipart upload with a "file" field (CSV bytes) and
// returns a preview of rows with categories applied from saved rules.
func (h *importHandler) parse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "multipart form required (max 5 MB)")
		return
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "field 'file' is required")
		return
	}
	defer f.Close()

	content, err := io.ReadAll(io.LimitReader(f, maxUploadBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read uploaded file")
		return
	}

	preview, err := h.imports.ParseCSV(r.Context(), middleware.UserIDFrom(r.Context()), content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

// listRules returns all saved merchant→category rules for the authenticated user.
func (h *importHandler) listRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.imports.GetMerchantRules(r.Context(), middleware.UserIDFrom(r.Context()))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

// commit saves new merchant rules and posts all rows as transactions.
func (h *importHandler) commit(w http.ResponseWriter, r *http.Request) {
	var req dto.ImportCommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.AccountID == "" {
		writeError(w, http.StatusBadRequest, "accountId is required")
		return
	}

	result, err := h.imports.Commit(r.Context(), middleware.UserIDFrom(r.Context()), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, result)
}
