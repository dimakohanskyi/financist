package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	financistdb "github.com/dimakohanskyi/financist/internal/db"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dimakohanskyi/financist/internal/configs"
	"github.com/dimakohanskyi/financist/internal/graph"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {

	if error := godotenv.Load(); error != nil {
		log.Fatal("failed to load envs")
	}

	logLevel := os.Getenv("LOG_LEVEL")

	if logLevel == "" {
		log.Fatal("log level not configured")
	}

	if error := configs.LogerSetUp(logLevel); error != nil {
		log.Fatal("failed to set up logger")
	}

	dbConnection, err := sql.Open("sqlite", "financist.db")
	if err != nil {
		log.Fatal("application couldn't create or connect to db")
	}
	if error := dbConnection.Ping(); error != nil {
		log.Fatal("DB is not created or accessible")
	}

	if error := financistdb.RunMigrations(dbConnection); error != nil {
		log.Fatal("application couldn't run migrations")
	}

	gqlPort := os.Getenv("GQL_API_PORT")
	if gqlPort == "" {
		gqlPort = "8080"
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

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

	http.Handle("/", playground.ApolloSandboxHandler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", gqlPort)
	log.Fatal(http.ListenAndServe(":"+gqlPort, nil))
}
