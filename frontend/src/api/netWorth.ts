import { fromCents, gql, toDate } from "@/api/client";
import type { CapitalRole, NetWorthRange, NetWorthSummary, RiskLevel } from "@/types";

const RANGES: Record<NetWorthRange, string> = {
  "30d": "last_30_days",
  "90d": "last_90_days",
  "1y": "last_year",
  "5y": "last_5_years",
  all: "all",
};

// Breakdown key -> [color, lucide icon]. Dataviz 8-slot order; don't reorder.
const CLASS_META: Record<string, [string, string]> = {
  cash: ["#2a78d6", "wallet"],
  investments: ["#eb6834", "trending-up"],
  real_estate: ["#eda100", "building-2"],
  vehicles: ["#e87ba4", "car"],
  precious_metals: ["#008300", "gem"],
  other: ["#4a3aa7", "package"],
};

// Status colors: good / neutral / critical.
const STATUS_COLOR = { good: "var(--success)", neutral: "var(--text-muted)", bad: "var(--danger)" };
const ROLE_COLOR: Record<CapitalRole, string> = { income: STATUS_COLOR.good, neutral: STATUS_COLOR.neutral, drain: STATUS_COLOR.bad };
const RISK_COLOR: Record<RiskLevel, string> = { low: STATUS_COLOR.good, medium: STATUS_COLOR.neutral, high: STATUS_COLOR.bad };

interface Part {
  key: string;
  name: string;
  amount: number;
  percent: number;
}

const PART = "key name amount percent";

export async function fetchNetWorthSummary(range: NetWorthRange): Promise<NetWorthSummary> {
  const data = await gql<{
    netWorth: {
      current: number;
      changeAmount: number;
      changePercent: number | null;
      series: { date: string; amount: number }[];
      breakdown: Part[];
      capitalRoles: { role: CapitalRole; totalValue: number; monthlyCashFlow: number; count: number }[];
      riskLevels: { riskLevel: RiskLevel; totalValue: number; percent: number; items: Part[] }[];
    };
  }>(
    `query($range: NetWorthRange!) { netWorth(range: $range) {
      current changeAmount changePercent series { date amount } breakdown { ${PART} }
      capitalRoles { role totalValue monthlyCashFlow count }
      riskLevels { riskLevel totalValue percent items { ${PART} } }
    } }`,
    { range: RANGES[range] }
  );
  const n = data.netWorth;
  const part = (p: Part) => ({ key: p.key, name: p.name, amount: fromCents(p.amount), percent: p.percent });
  return {
    range,
    current: fromCents(n.current),
    change_amount: fromCents(n.changeAmount),
    change_percent: n.changePercent,
    series: n.series.map((p) => ({ date: toDate(p.date), value: fromCents(p.amount) })),
    breakdown: n.breakdown.map((p) => {
      const [color, icon] = CLASS_META[p.key] ?? CLASS_META.other;
      return { ...part(p), color, icon };
    }),
    capital_roles: n.capitalRoles.map((r) => ({
      role: r.role,
      label: r.role,
      color: ROLE_COLOR[r.role],
      total_value: fromCents(r.totalValue),
      monthly_cash_flow: fromCents(r.monthlyCashFlow),
      count: r.count,
    })),
    risk_levels: n.riskLevels.map((r) => ({
      risk_level: r.riskLevel,
      label: r.riskLevel,
      color: RISK_COLOR[r.riskLevel],
      total_value: fromCents(r.totalValue),
      percent: r.percent,
      items: r.items.map(part),
    })),
  };
}
