package main

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dimakohanskyi/financist/internal/configs"
	financistdb "github.com/dimakohanskyi/financist/internal/db"
	"github.com/dimakohanskyi/financist/internal/repository"
	"github.com/dimakohanskyi/financist/internal/server"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// App holds what the frontend calls through Wails bindings (window.go.main.App).
type App struct {
	ctx     context.Context
	backups *repository.BackupRepo
}

// ExportBackup asks where to save and writes the backup there. The webview can't download files itself.
// Returns false when the user cancels.
func (a *App) ExportBackup() (bool, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "financist-backup-" + time.Now().Format("2006-01-02") + ".json",
		Filters:         []runtime.FileFilter{{DisplayName: "JSON", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return false, err
	}
	b, err := a.backups.Export(a.ctx)
	if err != nil {
		return false, err
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, data, 0o600)
}

// Confirm shows a native yes/no dialog: WKWebView has no window.confirm, so it would always return false.
func (a *App) Confirm(message string) bool {
	res, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type: runtime.QuestionDialog, Title: "Financist", Message: message,
		Buttons: []string{"Yes", "No"}, DefaultButton: "No", CancelButton: "No",
	})
	return err == nil && res == "Yes"
}

func main() {
	configs.LogerSetUp(os.Getenv("LOG_LEVEL"))

	// The app bundle is read-only, so data lives in ~/Library/Application Support/Financist.
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dir = filepath.Join(dir, "Financist")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		log.Fatal(err)
	}
	gormDB, err := financistdb.Open(filepath.Join(dir, "financist.db"))
	if err != nil {
		log.Fatalf("application couldn't open or migrate db: %v", err)
	}

	app := &App{backups: repository.NewBackupRepo(gormDB)}
	err = wails.Run(&options.App{
		Title:     "Financist",
		Width:     1280,
		Height:    820,
		MinWidth:  900,
		MinHeight: 600,
		// The API runs in-process: the webview's fetch("/query") reaches the handler without a network port.
		AssetServer: &assetserver.Options{Assets: assets, Handler: server.New(gormDB)},
		OnStartup:   func(ctx context.Context) { app.ctx = ctx },
		Bind:        []any{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
