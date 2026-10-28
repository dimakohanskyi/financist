import { ApiError } from "@/api/client";
import { t } from "@/lib/i18n";

export async function exportBackup(): Promise<void> {
  // The desktop webview can't download files, so Go shows a save dialog and writes it.
  if (window.go) {
    await window.go.main.App.ExportBackup();
    return;
  }
  // Plain browser dev: regular download.
  const response = await fetch("/backup");
  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }
  const url = URL.createObjectURL(await response.blob());
  const link = document.createElement("a");
  link.href = url;
  link.download = `financist-backup-${new Date().toISOString().slice(0, 10)}.json`;
  link.click();
  URL.revokeObjectURL(url);
}

export async function importBackup(file: File): Promise<void> {
  const text = await file.text();
  try {
    JSON.parse(text);
  } catch {
    throw new Error(t("backup.invalidFile"));
  }
  const response = await fetch("/backup", { method: "POST", headers: { "Content-Type": "application/json" }, body: text });
  if (!response.ok) {
    throw new ApiError(response.status, await response.text());
  }
}
