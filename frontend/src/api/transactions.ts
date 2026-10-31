import { ACCOUNT_FIELDS, mapAccount, type GqlAccount } from "@/api/accounts";
import { CATEGORY_FIELDS, mapCategory, type GqlCategory } from "@/api/categories";
import {
  blankToNull,
  compact,
  fromCents,
  gql,
  toCents,
  toDate,
  toId,
  toIdOrNull,
  toTime,
  toTimeOrNull,
} from "@/api/client";
import { mapTag } from "@/api/tags";
import type {
  Account,
  Category,
  Transaction,
  TransactionInput,
  TransactionPage,
  TransactionSplitInput,
  TransactionType,
} from "@/types";

export type TransactionSort = "date_desc" | "amount_desc" | "amount_asc";

export interface TransactionFilters {
  year?: number;
  month?: number;
  start_date?: string;
  end_date?: string;
  account_id?: number;
  category_id?: number;
  tag_id?: number;
  type?: string;
  search?: string;
  sort?: TransactionSort;
  page?: number;
  page_size?: number;
}

interface GqlTransaction {
  id: string;
  accountId: string;
  toAccountId: string | null;
  categoryId: string | null;
  type: TransactionType;
  amount: number;
  originalAmount: number | null;
  originalCurrency: string | null;
  description: string;
  merchant: string;
  notes: string;
  date: string;
  tags: { id: string; name: string }[];
  splits: { id: string; categoryId: string | null; amount: number; note: string }[];
}

const TX_FIELDS =
  "id accountId toAccountId categoryId type amount originalAmount originalCurrency description merchant notes date tags { id name } splits { id categoryId amount note }";

// Accounts and categories ride along in the same request so each transaction can carry
// its account and category objects, like the old REST payload did.
const LOOKUPS = `accounts(includeArchived: true) { ${ACCOUNT_FIELDS} } categories { ${CATEGORY_FIELDS} }`;

interface Lookups {
  accounts: GqlAccount[];
  categories: GqlCategory[];
}

function mapper({ accounts, categories }: Lookups) {
  const accountById = new Map<string, Account>(accounts.map((a) => [a.id, mapAccount(a)]));
  const categoryById = new Map<string, Category>(categories.map((c) => [c.id, mapCategory(c)]));
  const category = (id: string | null) => (id ? (categoryById.get(id) ?? null) : null);
  return (t: GqlTransaction): Transaction => ({
    id: toId(t.id),
    account_id: toId(t.accountId),
    category_id: toIdOrNull(t.categoryId),
    transfer_account_id: toIdOrNull(t.toAccountId),
    type: t.type,
    amount: fromCents(t.amount),
    original_amount: t.originalAmount === null ? null : fromCents(t.originalAmount),
    original_currency: t.originalCurrency,
    description: t.description,
    merchant: blankToNull(t.merchant),
    notes: blankToNull(t.notes),
    date: toDate(t.date),
    account: accountById.get(t.accountId)!,
    category: category(t.categoryId),
    tags: t.tags.map(mapTag),
    splits: t.splits.map((s) => ({
      id: toId(s.id),
      category_id: toIdOrNull(s.categoryId),
      category: category(s.categoryId),
      amount: fromCents(s.amount),
      note: blankToNull(s.note),
    })),
  });
}

const pad = (n: number) => String(n).padStart(2, "0");

/** year/month narrow to that month or year; explicit start/end dates win. */
function dateRange(f: TransactionFilters): { from?: string; to?: string } {
  let from: string | undefined;
  let to: string | undefined;
  if (f.year && f.month) {
    const lastDay = new Date(Date.UTC(f.year, f.month, 0)).getUTCDate();
    from = `${f.year}-${pad(f.month)}-01`;
    to = `${f.year}-${pad(f.month)}-${pad(lastDay)}`;
  } else if (f.year) {
    from = `${f.year}-01-01`;
    to = `${f.year}-12-31`;
  }
  from = f.start_date ?? from;
  to = f.end_date ?? to;
  return { from: from && toTime(from), to: to && toTime(to) };
}

export async function fetchTransactions(filters: TransactionFilters = {}): Promise<TransactionPage> {
  const page = filters.page ?? 1;
  const pageSize = filters.page_size ?? 50;
  const data = await gql<{ transactions: { items: GqlTransaction[]; total: number } } & Lookups>(
    `query($accountId: ID, $categoryId: ID, $tagId: ID, $type: TransactionType, $search: String, $from: Time, $to: Time, $sort: TransactionSort, $limit: Int, $offset: Int) {
      transactions(accountId: $accountId, categoryId: $categoryId, tagId: $tagId, type: $type, search: $search, from: $from, to: $to, sort: $sort, limit: $limit, offset: $offset) {
        items { ${TX_FIELDS} } total
      }
      ${LOOKUPS}
    }`,
    compact({
      accountId: filters.account_id,
      categoryId: filters.category_id,
      tagId: filters.tag_id,
      type: filters.type || undefined,
      search: filters.search || undefined,
      sort: filters.sort,
      limit: pageSize,
      offset: (page - 1) * pageSize,
      ...dateRange(filters),
    })
  );
  return {
    items: data.transactions.items.map(mapper(data)),
    total: data.transactions.total,
    page,
    page_size: pageSize,
  };
}

// Used by CSV import's duplicate check (see pages/CsvImportPage.tsx) — needs
// every transaction in the imported date range for one account, not just a
// page of them, to know what's already there. Capped at 25 pages (5000 rows
// at the API's max page_size) so a huge date range can't turn one import
// into an unbounded fetch loop; beyond that, duplicate detection just
// silently covers less of the range instead of hanging.
const MAX_DUPLICATE_CHECK_PAGES = 25;

export async function fetchAllTransactionsInRange(filters: TransactionFilters): Promise<Transaction[]> {
  const pageSize = 200;
  const items: Transaction[] = [];
  for (let page = 1; page <= MAX_DUPLICATE_CHECK_PAGES; page++) {
    const result = await fetchTransactions({ ...filters, page, page_size: pageSize });
    items.push(...result.items);
    if (items.length >= result.total || result.items.length < pageSize) break;
  }
  return items;
}

/** Full range of years to offer in the year picker, from the earliest
 * transaction through the current year (see backend for the "gap year"
 * rationale). */
export async function fetchTransactionYears() {
  const data = await gql<{ transactionYears: number[] }>(`{ transactionYears }`);
  return data.transactionYears;
}

const mapSplits = (splits: TransactionSplitInput[]) =>
  splits.map((s) => ({ categoryId: s.category_id, amount: toCents(s.amount), note: s.note ?? "" }));

function toCreateInput(input: TransactionInput) {
  return {
    accountId: input.account_id,
    toAccountId: input.type === "transfer" ? input.transfer_account_id : null,
    categoryId: input.category_id,
    type: input.type,
    amount: toCents(input.amount),
    currency: input.currency || null,
    description: input.description,
    merchant: input.merchant ?? "",
    notes: input.notes ?? "",
    date: toTime(input.date),
    tagIds: input.tag_ids ?? [],
    splits: mapSplits(input.splits ?? []),
  };
}

async function fetchOne(id: string): Promise<Transaction> {
  const data = await gql<{ transaction: GqlTransaction } & Lookups>(
    `query($id: ID!) { transaction(id: $id) { ${TX_FIELDS} } ${LOOKUPS} }`,
    { id }
  );
  return mapper(data)(data.transaction);
}

export async function createTransaction(input: TransactionInput) {
  const data = await gql<{ createTransaction: { id: string } }>(
    `mutation($input: CreateTransactionInput!) { createTransaction(input: $input) { id } }`,
    { input: toCreateInput(input) }
  );
  return fetchOne(data.createTransaction.id);
}

/** CSV import — see pages/CsvImportPage.tsx. All-or-nothing on the backend:
 * either every row is created, or (on a validation error) none are. */
export async function bulkCreateTransactions(items: TransactionInput[]) {
  const data = await gql<{ bulkCreateTransactions: number }>(
    `mutation($inputs: [CreateTransactionInput!]!) { bulkCreateTransactions(inputs: $inputs) }`,
    { inputs: items.map(toCreateInput) }
  );
  return { created: data.bulkCreateTransactions };
}

export async function updateTransaction(id: number, input: Partial<TransactionInput>) {
  const data = await gql<{ updateTransaction: { id: string } }>(
    `mutation($id: ID!, $input: UpdateTransactionInput!) { updateTransaction(id: $id, input: $input) { id } }`,
    {
      id,
      input: compact({
        accountId: input.account_id,
        // null means "unchanged" here; the backend drops it when the type stops being a transfer
        toAccountId: input.transfer_account_id ?? undefined,
        // null clears the category, so keep it when sent
        categoryId: input.category_id,
        type: input.type,
        amount: input.amount === undefined ? undefined : toCents(input.amount),
        currency: input.currency || undefined,
        description: input.description,
        merchant: input.merchant === undefined ? undefined : (input.merchant ?? ""),
        notes: input.notes === undefined ? undefined : (input.notes ?? ""),
        date: input.date === undefined ? undefined : toTimeOrNull(input.date),
        tagIds: input.tag_ids,
        splits: input.splits === undefined ? undefined : mapSplits(input.splits ?? []),
      }),
    }
  );
  return fetchOne(data.updateTransaction.id);
}

export async function deleteTransaction(id: number) {
  await gql(`mutation($id: ID!) { deleteTransaction(id: $id) }`, { id });
}
