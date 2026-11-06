import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// Desktop webview has no window.confirm (always false), so ask Go for a native dialog.
export async function confirmAction(message: string): Promise<boolean> {
  return window.go ? window.go.main.App.Confirm(message) : window.confirm(message);
}
