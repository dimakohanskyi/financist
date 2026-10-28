import { compact, fromCents, gql, toCents } from "@/api/client";
import type { AppSettings } from "@/types";
import { version } from "../../package.json";

interface GqlSettings {
  currency: string;
  negativeCashFlowThresholdMonths: number;
  netWorthDeclineThresholdMonths: number;
  riskyAllocationThresholdPercent: number;
  idleCashThresholdAmount: number;
  idleCashThresholdDays: number;
}

const FIELDS =
  "currency negativeCashFlowThresholdMonths netWorthDeclineThresholdMonths riskyAllocationThresholdPercent idleCashThresholdAmount idleCashThresholdDays";

const mapSettings = (s: GqlSettings): AppSettings => ({
  currency: s.currency,
  negative_cash_flow_threshold_months: s.negativeCashFlowThresholdMonths,
  net_worth_decline_threshold_months: s.netWorthDeclineThresholdMonths,
  risky_allocation_threshold_percent: s.riskyAllocationThresholdPercent,
  idle_cash_threshold_amount: fromCents(s.idleCashThresholdAmount),
  idle_cash_threshold_days: s.idleCashThresholdDays,
  app_version: version,
});

export async function fetchSettings() {
  const data = await gql<{ settings: GqlSettings }>(`{ settings { ${FIELDS} } }`);
  return mapSettings(data.settings);
}

// Partial — the backend applies whatever fields are present (PATCH
// semantics), so callers only send the field(s) they're changing.
export async function updateSettings(input: Partial<AppSettings>) {
  const data = await gql<{ updateSettings: GqlSettings }>(
    `mutation($input: UpdateSettingsInput!) { updateSettings(input: $input) { ${FIELDS} } }`,
    {
      input: compact({
        currency: input.currency,
        negativeCashFlowThresholdMonths: input.negative_cash_flow_threshold_months,
        netWorthDeclineThresholdMonths: input.net_worth_decline_threshold_months,
        riskyAllocationThresholdPercent: input.risky_allocation_threshold_percent,
        idleCashThresholdAmount:
          input.idle_cash_threshold_amount === undefined ? undefined : toCents(input.idle_cash_threshold_amount),
        idleCashThresholdDays: input.idle_cash_threshold_days,
      }),
    }
  );
  return mapSettings(data.updateSettings);
}
