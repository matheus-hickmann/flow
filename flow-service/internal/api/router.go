// Package api wires HTTP routes for the service.
// All handlers live here; commands/queries are pure business logic.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/hickmann/flow-service/internal/api/middleware"
	"github.com/hickmann/flow-service/internal/config"
	flowdynamo "github.com/hickmann/flow-service/internal/dynamodb"
	"github.com/hickmann/flow-service/internal/service"
)

// Deps groups everything a handler may need so we wire it once in main.
type Deps struct {
	Cfg    config.Config
	Dynamo flowdynamo.API
}

// NewRouter builds the chi router with all middleware and routes.
// Returns *chi.Mux (concrete) so it works both as http.Handler (server)
// and as the chiadapter input (Lambda).
func NewRouter(deps Deps) *chi.Mux {
	middleware.Setup(deps.Cfg.JWTSecret, deps.Cfg.DynamoDBEndpoint)

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{deps.Cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Origin", "Authorization"},
		AllowCredentials: true,
	}))

	// Health endpoint — public, used by Docker/load balancers.
	r.Get("/q/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "UP"})
	})

	authSvc := service.NewAuthService(deps.Dynamo, deps.Cfg.TableName)
	recoverySvc := service.NewRecoveryService(deps.Dynamo, deps.Cfg.TableName)
	accountSvc := service.NewAccountService(deps.Dynamo, deps.Cfg.TableName)
	systemSvc := service.NewSystemAccountsService(deps.Dynamo, deps.Cfg.TableName, accountSvc)
	txnSvc := service.NewTransactionService(deps.Dynamo, deps.Cfg.TableName, systemSvc)
	categorySvc := service.NewCategoryService(deps.Dynamo, deps.Cfg.TableName)
	planningSvc := service.NewPlanningService(deps.Dynamo, deps.Cfg.TableName)
	dashboardSvc := service.NewDashboardService(accountSvc, txnSvc, planningSvc)
	reportSvc := service.NewReportService(accountSvc, txnSvc, categorySvc)
	importSvc := service.NewImportService(deps.Dynamo, deps.Cfg.TableName, txnSvc)
	groupSvc := service.NewGroupService(deps.Dynamo, deps.Cfg.TableName, accountSvc)
	inviteSvc := service.NewInviteService(deps.Dynamo, deps.Cfg.TableName, groupSvc)
	debtSvc := service.NewDebtService(deps.Dynamo, deps.Cfg.TableName)

	r.Mount("/api/v1/auth", newAuthHandler(authSvc, deps.Cfg.DynamoDBEndpoint != "").routes())
	r.Mount("/api/v1/users", newUserHandler(authSvc, recoverySvc).routes())
	r.Mount("/api/v1/ledger", newLedgerHandler(accountSvc, txnSvc).routes())
	r.Mount("/api/v1/categories", newCategoryHandler(categorySvc).routes())
	r.Mount("/api/v1/planning", newPlanningHandler(planningSvc).routes())
	r.Mount("/api/v1/dashboard", newDashboardHandler(dashboardSvc).routes())
	r.Mount("/api/v1/reports", newReportHandler(reportSvc).routes())
	r.Mount("/api/v1/imports", newImportHandler(importSvc).routes())
	r.Mount("/api/v1/groups", newGroupHandler(groupSvc, inviteSvc, authSvc).routes())
	r.Mount("/api/v1/invites", newInviteHandler(inviteSvc, authSvc).routes())
	r.Mount("/api/v1/debts", newDebtHandler(debtSvc).routes())

	return r
}

// writeJSON is a small helper to keep handlers terse.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
