import { compact, fromCents, gql, toId } from "@/api/client";
import { getCurrency } from "@/lib/i18n";
import type { Account, AccountInput, AccountType, AccountWithBalance } from "@/types";

export interface GqlAccount {
  id: string;
  name: string;
  accountType: AccountType;
  currency: string;
  color: string | null;
  isArchived: boolean;
}

export const ACCOUNT_FIELDS = "id name accountType currency color isArchived";
const FIELDS = `${ACCOUNT_FIELDS} balance`;

export const mapAccount = (a: GqlAccount): Account => ({
  id: toId(a.id),
  name: a.name,
  type: a.accountType,
  currency: a.currency,
  color: a.color,
  is_archived: a.isArchived,
});

const mapWithBalance = (a: GqlAccount & { balance: number }): AccountWithBalance => ({
  ...mapAccount(a),
  balance: fromCents(a.balance),
});

export async function fetchAccounts(includeArchived = false) {
  const data = await gql<{ accounts: (GqlAccount & { balance: number })[] }>(
    `query($includeArchived: Boolean) { accounts(includeArchived: $includeArchived) { ${FIELDS} } }`,
    { includeArchived }
  );
  return data.accounts.map(mapWithBalance);
}

export async function createAccount(input: AccountInput) {
  // ponytail: new accounts take the display currency; the form has no currency picker
  const data = await gql<{ createAccount: GqlAccount & { balance: number } }>(
    `mutation($input: CreateAccountInput!) { createAccount(input: $input) { ${FIELDS} } }`,
    { input: { name: input.name, accountType: input.type, currency: getCurrency() } }
  );
  return mapWithBalance(data.createAccount);
}

export async function updateAccount(id: number, input: Partial<AccountInput> & { is_archived?: boolean }) {
  const data = await gql<{ updateAccount: GqlAccount & { balance: number } }>(
    `mutation($id: ID!, $input: UpdateAccountInput!) { updateAccount(id: $id, input: $input) { ${FIELDS} } }`,
    { id, input: compact({ name: input.name, accountType: input.type, isArchived: input.is_archived }) }
  );
  return mapWithBalance(data.updateAccount);
}

export async function deleteAccount(id: number) {
  await gql(`mutation($id: ID!) { deleteAccount(id: $id) }`, { id });
}
