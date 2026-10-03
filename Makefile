run:
	go run cmd/fin_app.go
lint:
	golangci-lint run
gen:
	sqlc generate
	go generate ./internal/graph
