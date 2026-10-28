import { fromCents, gql, toCents, toId } from "@/api/client";
import type { Budget, BudgetInput, BudgetStatusResponse } from "@/types";

interface GqlBudget {
  id: string;
  categoryId: string;
  category: { name: string; color: string; icon: string | null };
  monthlyLimit: number;
}

const FIELDS = "id categoryId category { name color icon } monthlyLimit";

const mapBudget = (b: GqlBudget): Budget => ({
  id: toId(b.id),
  category_id: toId(b.categoryId),
  category_name: b.category.name,
  category_color: b.category.color,
  category_icon: b.category.icon,
  monthly_limit: fromCents(b.monthlyLimit),
});

export async function fetchBudgets() {
  const data = await gql<{ budgets: GqlBudget[] }>(`{ budgets { ${FIELDS} } }`);
  return data.budgets.map(mapBudget);
}

export async function fetchBudgetStatus(year: number, month: number): Promise<BudgetStatusResponse> {
  const data = await gql<{
    budgetStatus: { budget: GqlBudget; spent: number; remaining: number; percent: number; isOverBudget: boolean }[];
  }>(
    `query($year: Int, $month: Int) { budgetStatus(year: $year, month: $month) { budget { ${FIELDS} } spent remaining percent isOverBudget } }`,
    { year, month }
  );
  return {
    year,
    month,
    items: data.budgetStatus.map((s) => {
      const { id, ...budget } = mapBudget(s.budget);
      return {
        ...budget,
        budget_id: id,
        spent: fromCents(s.spent),
        remaining: fromCents(s.remaining),
        percent: s.percent,
        is_over_budget: s.isOverBudget,
      };
    }),
  };
}

export async function createBudget(input: BudgetInput) {
  const data = await gql<{ createBudget: GqlBudget }>(
    `mutation($categoryId: ID!, $monthlyLimit: Int64!) { createBudget(categoryId: $categoryId, monthlyLimit: $monthlyLimit) { ${FIELDS} } }`,
    { categoryId: input.category_id, monthlyLimit: toCents(input.monthly_limit) }
  );
  return mapBudget(data.createBudget);
}

export async function updateBudget(id: number, monthlyLimit: string) {
  const data = await gql<{ updateBudget: GqlBudget }>(
    `mutation($id: ID!, $monthlyLimit: Int64!) { updateBudget(id: $id, monthlyLimit: $monthlyLimit) { ${FIELDS} } }`,
    { id, monthlyLimit: toCents(monthlyLimit) }
  );
  return mapBudget(data.updateBudget);
}

export async function deleteBudget(id: number) {
  await gql(`mutation($id: ID!) { deleteBudget(id: $id) }`, { id });
}
