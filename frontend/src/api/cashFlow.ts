import { fromCents, gql, toDateOrNull, toTimeOrNull } from "@/api/client";
import type { CashFlowResponse } from "@/types";

export async function fetchCashFlow(startDate?: string, endDate?: string): Promise<CashFlowResponse> {
  const data = await gql<{
    cashFlow: {
      from: string | null;
      to: string | null;
      months: { year: number; month: number; income: number; expense: number; net: number }[];
      totalIncome: number;
      totalExpense: number;
      totalNet: number;
    };
  }>(
    `query($from: Time, $to: Time) { cashFlow(from: $from, to: $to) { from to months { year month income expense net } totalIncome totalExpense totalNet } }`,
    { from: toTimeOrNull(startDate), to: toTimeOrNull(endDate) }
  );
  const c = data.cashFlow;
  return {
    start_date: toDateOrNull(c.from),
    end_date: toDateOrNull(c.to),
    points: c.months.map((m) => ({
      year: m.year,
      month: m.month,
      income: fromCents(m.income),
      expense: fromCents(m.expense),
      net: fromCents(m.net),
    })),
    total_income: fromCents(c.totalIncome),
    total_expense: fromCents(c.totalExpense),
    total_net: fromCents(c.totalNet),
  };
}
