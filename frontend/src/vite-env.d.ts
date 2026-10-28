/// <reference types="vite/client" />

interface Window {
  /** Wails bindings; absent when running in a plain browser. */
  go?: { main: { App: { ExportBackup(): Promise<boolean> } } };
}
