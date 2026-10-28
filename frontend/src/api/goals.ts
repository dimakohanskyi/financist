import { compact, fromCents, gql, toCents, toDateOrNull, toId, toTime, toTimeOrNull } from "@/api/client";
import type { Goal, GoalContributionInput, GoalInput } from "@/types";

interface GqlGoal {
  id: string;
  name: string;
  targetAmount: number;
  targetDate: string | null;
  currentAmount: number;
  remaining: number;
  percent: number;
  isReached: boolean;
}

const FIELDS = "id name targetAmount targetDate currentAmount remaining percent isReached";

const mapGoal = (g: GqlGoal): Goal => ({
  id: toId(g.id),
  name: g.name,
  target_amount: fromCents(g.targetAmount),
  target_date: toDateOrNull(g.targetDate),
  current_amount: fromCents(g.currentAmount),
  remaining: fromCents(g.remaining),
  percent: g.percent,
  is_reached: g.isReached,
});

export async function fetchGoals() {
  const data = await gql<{ goals: GqlGoal[] }>(`{ goals { ${FIELDS} } }`);
  return data.goals.map(mapGoal);
}

export async function createGoal(input: GoalInput) {
  const data = await gql<{ createGoal: GqlGoal }>(
    `mutation($input: CreateGoalInput!) { createGoal(input: $input) { ${FIELDS} } }`,
    { input: { name: input.name, targetAmount: toCents(input.target_amount), targetDate: toTimeOrNull(input.target_date) } }
  );
  return mapGoal(data.createGoal);
}

export async function updateGoal(id: number, input: Partial<GoalInput>) {
  const data = await gql<{ updateGoal: GqlGoal }>(
    `mutation($id: ID!, $input: UpdateGoalInput!) { updateGoal(id: $id, input: $input) { ${FIELDS} } }`,
    {
      id,
      input: compact({
        name: input.name,
        targetAmount: input.target_amount === undefined ? undefined : toCents(input.target_amount),
        // null clears the date
        targetDate: input.target_date === undefined ? undefined : toTimeOrNull(input.target_date),
      }),
    }
  );
  return mapGoal(data.updateGoal);
}

export async function deleteGoal(id: number) {
  await gql(`mutation($id: ID!) { deleteGoal(id: $id) }`, { id });
}

export async function addGoalContribution(id: number, input: GoalContributionInput) {
  const data = await gql<{ addGoalContribution: GqlGoal }>(
    `mutation($goalId: ID!, $amount: Int64!, $date: Time!, $note: String) { addGoalContribution(goalId: $goalId, amount: $amount, date: $date, note: $note) { ${FIELDS} } }`,
    { goalId: id, amount: toCents(input.amount), date: toTime(input.date), note: input.note ?? "" }
  );
  return mapGoal(data.addGoalContribution);
}
