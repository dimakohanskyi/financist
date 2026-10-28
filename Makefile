run:
	go run cmd/fin_app.go
lint:
	golangci-lint run
gen:
	go generate ./internal/graph
# Desktop app with live reload
dev:
	wails dev
# build/bin/Financist.app
app:
	wails build -clean
