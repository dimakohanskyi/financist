import { fromCents, gql, toDateOrNull, toTimeOrNull } from "@/api/client";
import { CATEGORY_AMOUNT_FIELDS, mapCategoryAmount, type GqlCategoryAmount } from "@/api/dashboard";
import type { CategoryKind, CategoryRankingReport, CategorySpendingReport } from "@/types";

export async function fetchCategorySpendingReport(
  categoryId: number,
  startDate?: string,
  endDate?: string
): Promise<CategorySpendingReport> {
  const data = await gql<{
    categorySpending: {
      categoryId: string;
      category: { name: string; color: string; icon: string | null };
      from: string | null;
      to: string | null;
      total: number;
      transactionCount: number;
      averagePerMonth: number;
      months: { year: number; month: number; amount: number }[];
    };
  }>(
    `query($categoryId: ID!, $from: Time, $to: Time) { categorySpending(categoryId: $categoryId, from: $from, to: $to) {
      categoryId category { name color icon } from to total transactionCount averagePerMonth months { year month amount }
    } }`,
    { categoryId, from: toTimeOrNull(startDate), to: toTimeOrNull(endDate) }
  );
  const r = data.categorySpending;
  return {
    category_id: Number(r.categoryId),
    category_name: r.category.name,
    category_color: r.category.color,
    category_icon: r.category.icon,
    start_date: toDateOrNull(r.from),
    end_date: toDateOrNull(r.to),
    total_amount: fromCents(r.total),
    transaction_count: r.transactionCount,
    average_per_month: fromCents(r.averagePerMonth),
    series: r.months.map((m) => ({ year: m.year, month: m.month, amount: fromCents(m.amount) })),
  };
}

export async function fetchCategoryRanking(
  kind: CategoryKind,
  startDate?: string,
  endDate?: string
): Promise<CategoryRankingReport> {
  const data = await gql<{
    categoryRanking: { from: string | null; to: string | null; total: number; items: GqlCategoryAmount[] };
  }>(
    `query($kind: CategoryKind!, $from: Time, $to: Time) { categoryRanking(kind: $kind, from: $from, to: $to) {
      from to total items { ${CATEGORY_AMOUNT_FIELDS} }
    } }`,
    { kind, from: toTimeOrNull(startDate), to: toTimeOrNull(endDate) }
  );
  const r = data.categoryRanking;
  return {
    start_date: toDateOrNull(r.from),
    end_date: toDateOrNull(r.to),
    total_amount: fromCents(r.total),
    items: r.items.map((item) => {
      const mapped = mapCategoryAmount(item);
      return { ...mapped, category_id: mapped.category_id ?? 0 };
    }),
  };
}
