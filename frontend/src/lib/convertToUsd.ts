// Totals are converted to USD on the backend (X-Convert-USD) because summing
// accounts in different currencies is meaningless. Only the dashboard turns
// it off, to show one non-USD account in its own currency.
let on = true;
let nativeCurrency: string | undefined;

export const isConvertToUsd = () => on;

/** native: the currency amounts come back in while conversion is off. */
export function setConvertToUsd(next: boolean, native?: string) {
  on = next;
  nativeCurrency = next ? undefined : native;
}

export const getNativeCurrency = () => nativeCurrency;
