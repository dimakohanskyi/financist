import { isConvertToUsd } from "@/lib/convertToUsd";

// Talks GraphQL to the Go backend, served in-process by Wails at /query.
// The api/* modules map its shapes (camelCase, minor units, string IDs, RFC3339 times)
// onto the frontend types (snake_case, decimal strings, numeric IDs, YYYY-MM-DD dates),
// so components never see the wire format.

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function gql<T>(query: string, variables?: Record<string, unknown>): Promise<T> {
  const response = await fetch("/query", {
    method: "POST",
    headers: { "Content-Type": "application/json", ...(isConvertToUsd() ? { "X-Convert-USD": "1" } : {}) },
    body: JSON.stringify({ query, variables }),
  });
  if (!response.ok) {
    throw new ApiError(response.status, (await response.text()) || response.statusText);
  }
  const body = (await response.json()) as { data?: T; errors?: { message: string }[] };
  if (body.errors?.length) {
    throw new ApiError(400, body.errors.map((e) => e.message).join("; "));
  }
  return body.data as T;
}

// Amounts are cents on the wire, decimal strings ("12.34") in the UI.
export const toCents = (amount: string | number) => Math.round(Number(amount) * 100);
export const fromCents = (cents: number) => (cents / 100).toFixed(2);
export const fromCentsOrNull = (cents: number | null | undefined) => (cents == null ? null : fromCents(cents));

export const toId = (id: string) => Number(id);
export const toIdOrNull = (id: string | null | undefined) => (id == null ? null : Number(id));

// Dates are YYYY-MM-DD in the UI, RFC3339 on the wire.
export const toTime = (date: string) => (date.length === 10 ? `${date}T00:00:00Z` : date);
export const toTimeOrNull = (date: string | null | undefined) => (date ? toTime(date) : null);
export const toDate = (time: string) => time.slice(0, 10);
export const toDateOrNull = (time: string | null | undefined) => (time ? toDate(time) : null);

// "" and null both mean "no value" in the old REST forms.
export const blankToNull = (s: string | null | undefined) => (s ? s : null);

/** Drops undefined keys so partial updates leave unsent fields untouched. */
export function compact<T extends Record<string, unknown>>(obj: T): Partial<T> {
  return Object.fromEntries(Object.entries(obj).filter(([, v]) => v !== undefined)) as Partial<T>;
}
