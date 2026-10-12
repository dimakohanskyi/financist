run:
	go run cmd/fin_app.go
lint:
	golangci-lint run
gen:
	go generate ./internal/graph
