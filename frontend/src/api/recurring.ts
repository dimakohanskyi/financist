import { ACCOUNT_FIELDS, type GqlAccount } from "@/api/accounts";
import { CATEGORY_FIELDS, type GqlCategory } from "@/api/categories";
import { blankToNull, compact, fromCents, gql, toCents, toDate, toDateOrNull, toId, toIdOrNull, toTime } from "@/api/client";
import type { RecurringFrequency, RecurringTransaction, RecurringTransactionInput, TransactionType } from "@/types";

interface GqlRecurring {
  id: string;
  accountId: string;
  toAccountId: string | null;
  categoryId: string | null;
  type: TransactionType;
  amount: number;
  description: string;
  merchant: string;
  notes: string;
  frequency: RecurringFrequency;
  anchorDate: string;
  lastPostedDate: string | null;
  isActive: boolean;
  nextDueDate: string;
  isDue: boolean;
  daysUntilDue: number;
}

const FIELDS =
  "id accountId toAccountId categoryId type amount description merchant notes frequency anchorDate lastPostedDate isActive nextDueDate isDue daysUntilDue";
const LOOKUPS = `accounts(includeArchived: true) { ${ACCOUNT_FIELDS} } categories { ${CATEGORY_FIELDS} }`;

interface Lookups {
  accounts: GqlAccount[];
  categories: GqlCategory[];
}

function mapper({ accounts, categories }: Lookups) {
  const accountName = new Map(accounts.map((a) => [a.id, a.name]));
  const categoryById = new Map(categories.map((c) => [c.id, c]));
  return (r: GqlRecurring): RecurringTransaction => {
    const category = r.categoryId ? categoryById.get(r.categoryId) : undefined;
    return {
      id: toId(r.id),
      account_id: toId(r.accountId),
      account_name: accountName.get(r.accountId) ?? "",
      category_id: toIdOrNull(r.categoryId),
      category_name: category?.name ?? null,
      category_color: category?.color ?? null,
      category_icon: category?.icon ?? null,
      transfer_account_id: toIdOrNull(r.toAccountId),
      transfer_account_name: r.toAccountId ? (accountName.get(r.toAccountId) ?? null) : null,
      type: r.type,
      amount: fromCents(r.amount),
      description: r.description,
      merchant: blankToNull(r.merchant),
      notes: blankToNull(r.notes),
      frequency: r.frequency,
      anchor_date: toDate(r.anchorDate),
      last_posted_date: toDateOrNull(r.lastPostedDate),
      is_active: r.isActive,
      next_due_date: toDate(r.nextDueDate),
      is_due: r.isDue,
      days_until_due: r.daysUntilDue,
    };
  };
}

export async function fetchRecurring() {
  const data = await gql<{ recurringTransactions: GqlRecurring[] } & Lookups>(
    `{ recurringTransactions { ${FIELDS} } ${LOOKUPS} }`
  );
  return data.recurringTransactions.map(mapper(data));
}

// Mutations return only the id, then re-read the list for the joined names.
async function fetchOne(id: string) {
  return (await fetchRecurring()).find((r) => r.id === toId(id))!;
}

export async function createRecurring(input: RecurringTransactionInput) {
  const data = await gql<{ createRecurringTransaction: { id: string } }>(
    `mutation($input: CreateRecurringTransactionInput!) { createRecurringTransaction(input: $input) { id } }`,
    {
      input: {
        accountId: input.account_id,
        toAccountId: input.type === "transfer" ? input.transfer_account_id : null,
        categoryId: input.category_id,
        type: input.type,
        amount: toCents(input.amount),
        description: input.description,
        merchant: input.merchant ?? "",
        frequency: input.frequency,
        anchorDate: toTime(input.anchor_date),
      },
    }
  );
  return fetchOne(data.createRecurringTransaction.id);
}

export async function updateRecurring(id: number, input: Partial<RecurringTransactionInput>) {
  const data = await gql<{ updateRecurringTransaction: { id: string } }>(
    `mutation($id: ID!, $input: UpdateRecurringTransactionInput!) { updateRecurringTransaction(id: $id, input: $input) { id } }`,
    {
      id,
      input: compact({
        accountId: input.account_id,
        toAccountId: input.transfer_account_id ?? undefined,
        categoryId: input.category_id,
        type: input.type,
        amount: input.amount === undefined ? undefined : toCents(input.amount),
        description: input.description,
        merchant: input.merchant === undefined ? undefined : (input.merchant ?? ""),
        frequency: input.frequency,
        anchorDate: input.anchor_date === undefined ? undefined : toTime(input.anchor_date),
      }),
    }
  );
  return fetchOne(data.updateRecurringTransaction.id);
}

export async function deleteRecurring(id: number) {
  await gql(`mutation($id: ID!) { deleteRecurringTransaction(id: $id) }`, { id });
}

export async function postRecurring(id: number) {
  await gql(`mutation($id: ID!) { postRecurringTransaction(id: $id) { id } }`, { id });
  return fetchOne(String(id));
}
