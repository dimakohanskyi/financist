import { gql } from "@/api/client";
import type { FinancialAlert } from "@/types";

export async function fetchFinancialAlerts() {
  const data = await gql<{ alerts: { key: string; level: string; params: Record<string, number> }[] }>(
    `{ alerts { key level params } }`
  );
  return { alerts: data.alerts.map((a): FinancialAlert => ({ key: a.key, severity: a.level, params: a.params })) };
}
