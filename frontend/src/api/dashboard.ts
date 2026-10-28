import { fromCents, gql } from "@/api/client";
import type { CategoryBreakdownItem, DashboardSummary } from "@/types";

export interface GqlCategoryAmount {
  categoryId: string | null;
  category: { name: string; color: string; icon: string | null } | null;
  amount: number;
  percent: number;
  transactionCount: number;
  children: GqlCategoryAmount[];
}

const ONE = "categoryId category { name color icon } amount percent transactionCount";
export const CATEGORY_AMOUNT_FIELDS = `${ONE} children { ${ONE} }`;

// The slice folding everything past the top few has no category.
const OTHER = { name: "Other", color: "#898781", icon: "more-horizontal" };

export function mapCategoryAmount(c: GqlCategoryAmount): CategoryBreakdownItem & { transaction_count: number } {
  const meta = c.category ?? OTHER;
  return {
    category_id: c.categoryId == null ? null : Number(c.categoryId),
    name: meta.name,
    color: meta.color,
    icon: meta.icon,
    amount: fromCents(c.amount),
    percent: c.percent,
    transaction_count: c.transactionCount,
    children: c.children.map((child) => ({
      category_id: Number(child.categoryId),
      name: child.category?.name ?? OTHER.name,
      color: child.category?.color ?? OTHER.color,
      icon: child.category?.icon ?? null,
      amount: fromCents(child.amount),
    })),
  };
}

/** accountId narrows to that account; omitted means every account. */
export async function fetchDashboardSummary(year: number, month: number, accountId?: number): Promise<DashboardSummary> {
  const data = await gql<{
    dashboard: { year: number; month: number; income: number; spent: number; net: number; transferredOut: number; spendingByCategory: GqlCategoryAmount[] };
  }>(
    `query($year: Int, $month: Int, $accountId: ID) { dashboard(year: $year, month: $month, accountId: $accountId) { year month income spent net transferredOut spendingByCategory { ${CATEGORY_AMOUNT_FIELDS} } } }`,
    { year, month, accountId }
  );
  const d = data.dashboard;
  return {
    year: d.year,
    month: d.month,
    real_income: fromCents(d.income),
    spent: fromCents(d.spent),
    net: fromCents(d.net),
    transferred_out: fromCents(d.transferredOut),
    spending_by_category: d.spendingByCategory.map(mapCategoryAmount),
  };
}
