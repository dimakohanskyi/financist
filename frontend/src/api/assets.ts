import { blankToNull, compact, fromCents, fromCentsOrNull, gql, toCents, toDate, toId, toTimeOrNull } from "@/api/client";
import type { Asset, AssetClass, AssetInput, AssetUpdateInput, AssetValuationInput, CapitalRole, RiskLevel } from "@/types";

interface GqlAsset {
  id: string;
  name: string;
  assetClass: AssetClass;
  currency: string;
  notes: string;
  capitalRole: CapitalRole;
  monthlyCashFlow: number | null;
  riskLevel: RiskLevel;
  currentValue: number;
  asOfDate: string | null;
}

const FIELDS = "id name assetClass currency notes capitalRole monthlyCashFlow riskLevel currentValue asOfDate";

const mapAsset = (a: GqlAsset): Asset => ({
  id: toId(a.id),
  name: a.name,
  asset_class: a.assetClass,
  currency: a.currency,
  notes: blankToNull(a.notes),
  capital_role: a.capitalRole,
  monthly_cash_flow: fromCentsOrNull(a.monthlyCashFlow),
  risk_level: a.riskLevel,
  current_value: fromCents(a.currentValue),
  as_of_date: a.asOfDate ? toDate(a.asOfDate) : "",
});

const centsOrNull = (s: string | null | undefined) => (s == null || s === "" ? null : toCents(s));

export async function fetchAssets() {
  const data = await gql<{ assets: GqlAsset[] }>(`{ assets { ${FIELDS} } }`);
  return data.assets.map(mapAsset);
}

export async function createAsset(input: AssetInput) {
  const data = await gql<{ createAsset: GqlAsset }>(
    `mutation($input: CreateAssetInput!) { createAsset(input: $input) { ${FIELDS} } }`,
    {
      input: compact({
        name: input.name,
        assetClass: input.asset_class,
        currency: input.currency,
        notes: input.notes ?? "",
        capitalRole: input.capital_role,
        monthlyCashFlow: centsOrNull(input.monthly_cash_flow),
        riskLevel: input.risk_level,
        value: toCents(input.value),
        asOfDate: toTimeOrNull(input.as_of_date),
      }),
    }
  );
  return mapAsset(data.createAsset);
}

export async function updateAsset(id: number, input: AssetUpdateInput) {
  const data = await gql<{ updateAsset: GqlAsset }>(
    `mutation($id: ID!, $input: UpdateAssetInput!) { updateAsset(id: $id, input: $input) { ${FIELDS} } }`,
    {
      id,
      input: compact({
        name: input.name,
        assetClass: input.asset_class,
        notes: input.notes === undefined ? undefined : (input.notes ?? ""),
        capitalRole: input.capital_role,
        // null clears it
        monthlyCashFlow: input.monthly_cash_flow === undefined ? undefined : centsOrNull(input.monthly_cash_flow),
        riskLevel: input.risk_level,
      }),
    }
  );
  return mapAsset(data.updateAsset);
}

export async function addAssetValuation(id: number, input: AssetValuationInput) {
  const data = await gql<{ addAssetValuation: GqlAsset }>(
    `mutation($assetId: ID!, $value: Int64!, $asOfDate: Time) { addAssetValuation(assetId: $assetId, value: $value, asOfDate: $asOfDate) { ${FIELDS} } }`,
    { assetId: id, value: toCents(input.value), asOfDate: toTimeOrNull(input.as_of_date) }
  );
  return mapAsset(data.addAssetValuation);
}

export async function deleteAsset(id: number) {
  await gql(`mutation($id: ID!) { deleteAsset(id: $id) }`, { id });
}
