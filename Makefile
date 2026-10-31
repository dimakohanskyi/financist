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

# build and replace the app in /Applications
install: app
	rm -rf /Applications/Financist.app
	cp -R build/bin/Financist.app /Applications/
