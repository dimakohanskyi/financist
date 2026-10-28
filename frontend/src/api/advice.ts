import { gql } from "@/api/client";
import type { AdviceItem, AdviceResponse } from "@/types";

// Params holding money come back in cents.
const MONEY_PARAMS = ["amount", "current", "average"];

export async function fetchAdvice(): Promise<AdviceResponse> {
  const data = await gql<{ advice: { key: string; level: AdviceItem["tone"]; params: AdviceItem["params"] }[] }>(
    `{ advice { key level params } }`
  );
  return {
    items: data.advice.map((a) => ({
      key: a.key,
      tone: a.level,
      params: Object.fromEntries(
        Object.entries(a.params).map(([k, v]) => [k, MONEY_PARAMS.includes(k) ? Number(v) / 100 : v])
      ),
    })),
  };
}
