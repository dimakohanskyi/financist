package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dimakohanskyi/financist/internal/configs"
	financistdb "github.com/dimakohanskyi/financist/internal/db"
	"github.com/dimakohanskyi/financist/internal/server"
	"github.com/joho/godotenv"
)

// Dev server: the API over HTTP plus a playground. The desktop app is main.go at the repo root.
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

	gormDB, err := financistdb.Open("financist.db")
	if err != nil {
		log.Fatalf("application couldn't open or migrate db: %v", err)
	}

	gqlPort := os.Getenv("GQL_API_PORT")
	if gqlPort == "" {
		gqlPort = "8080"
	}

	mux := server.New(gormDB)
	mux.Handle("/", playground.ApolloSandboxHandler("GraphQL playground", "/query"))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", gqlPort)
	log.Fatal(http.ListenAndServe(":"+gqlPort, mux))
}
