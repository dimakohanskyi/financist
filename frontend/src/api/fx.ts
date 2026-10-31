import { gql } from "@/api/client";
import type { FxStatus } from "@/types";

interface GqlFxStatus {
  source: string;
  updatedAt: string | null;
  rates: { quote: string; rate: number; date: string }[];
}

const FIELDS = "source updatedAt rates { quote rate date }";

const mapFxStatus = (s: GqlFxStatus): FxStatus => ({ source: s.source, updated_at: s.updatedAt, rates: s.rates });

export async function fetchFxStatus() {
  const data = await gql<{ fxStatus: GqlFxStatus }>(`{ fxStatus { ${FIELDS} } }`);
  return mapFxStatus(data.fxStatus);
}

export async function refreshFxRates() {
  const data = await gql<{ refreshFxRates: GqlFxStatus }>(`mutation { refreshFxRates { ${FIELDS} } }`);
  return mapFxStatus(data.refreshFxRates);
}
