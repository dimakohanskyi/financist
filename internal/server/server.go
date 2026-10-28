package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/dimakohanskyi/financist/internal/graph"
	"github.com/dimakohanskyi/financist/internal/repository"
	"github.com/dimakohanskyi/financist/internal/service"
	"github.com/vektah/gqlparser/v2/ast"
	"gorm.io/gorm"
)

// New serves the GraphQL API at /query and backups at /backup.
// Shared by the dev server (cmd) and the desktop app (main.go).
func New(gormDB *gorm.DB) *http.ServeMux {
	accountRepo := repository.NewAccountRepo(gormDB)
	categoryRepo := repository.NewCategoryRepo(gormDB)
	txRepo := repository.NewTransactionRepo(gormDB)
	txService := service.NewTransactionService(txRepo, accountRepo, categoryRepo)
	assetRepo := repository.NewAssetRepo(gormDB)
	settingsRepo := repository.NewSettingsRepo(gormDB)
	budgetService := service.NewBudgetService(repository.NewBudgetRepo(gormDB), categoryRepo, txRepo)

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		AccountRepo:  accountRepo,
		TxService:    txService,
		CategorySvc:  service.NewCategoryService(categoryRepo),
		TagRepo:      repository.NewTagRepo(gormDB),
		Recurring:    service.NewRecurringService(repository.NewRecurringRepo(gormDB), txService),
		BudgetSvc:    budgetService,
		GoalRepo:     repository.NewGoalRepo(gormDB),
		AssetRepo:    assetRepo,
		SettingsRepo: settingsRepo,
		Reports:      service.NewReportService(txRepo, accountRepo, categoryRepo, assetRepo, settingsRepo, budgetService),
	}}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	// SetQueryCache caches parsed GraphQL queries to avoid re-parsing
	// When a client sends a GraphQL query string, it must be parsed into an AST (Abstract Syntax Tree)
	// to understand which fields to resolve. This parsing is expensive.
	// LRU cache stores up to 1000 parsed queries in memory. When the same query is sent again,
	// we use the cached AST instead of re-parsing, which can be 10x-100x faster.
	// LRU = Least Recently Used: when cache is full, oldest unused query is removed.
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Introspection allows clients to query the schema structure (used by tools like GraphQL Playground)
	srv.Use(extension.Introspection{})

	// AutomaticPersistedQuery (APQ) allows clients to send only a hash of the query instead of the full query string
	// Reduces network bandwidth. Cache stores 100 APQ hashes in memory.
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	mux := http.NewServeMux()
	mux.Handle("/query", srv)

	// Backups are plain HTTP: a file to download and upload, not a query.
	backups := repository.NewBackupRepo(gormDB)
	mux.HandleFunc("GET /backup", func(w http.ResponseWriter, r *http.Request) {
		b, err := backups.Export(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", `attachment; filename="financist-backup.json"`)
		json.NewEncoder(w).Encode(b)
	})
	// Replaces all data with the uploaded backup, all or nothing.
	mux.HandleFunc("POST /backup", func(w http.ResponseWriter, r *http.Request) {
		var b repository.Backup
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 100<<20)).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("restoring backup exported at %s: %d accounts, %d transactions", b.ExportedAt, len(b.Accounts), len(b.Transactions))
		if err := backups.Import(r.Context(), &b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return mux
}
