import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { formatCurrency, formatTransactionDate } from "@/lib/format";
import { useTranslation } from "@/lib/i18n";

interface DailySpendingChartProps {
  year: number;
  month: number;
  /** Spent per day, day 1 first. */
  daily: number[];
  /** Day of the month to mark, when the shown month is the current one. */
  today?: number;
  loading?: boolean;
}

// Gridlines at the top, middle and baseline.
const GRID = ["top-0", "top-1/2", "bottom-0"];

const pad = (n: number) => String(n).padStart(2, "0");

export function DailySpendingChart({
  year,
  month,
  daily,
  today,
  loading,
}: DailySpendingChartProps) {
  const { t } = useTranslation();
  // The webview has no native title tooltips, so hover is tracked here.
  const [hover, setHover] = useState<number | null>(null);
  const max = Math.max(0, ...daily);
  const total = daily.reduce((sum, v) => sum + v, 0);
  // days elapsed, so a month in progress is not averaged over days not yet lived
  const elapsed = Math.max(1, today ?? daily.length);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-medium normal-case tracking-normal text-text-secondary">
          {t("dashboard.dailySpendingTitle")}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="h-44 motion-safe:animate-pulse rounded bg-surface-2" />
        ) : total === 0 ? (
          <p className="py-10 text-center text-sm text-text-muted">
            {t("dashboard.noExpensesThisMonth")}
          </p>
        ) : (
          <>
            <div
              className="relative mt-8 h-44"
              role="img"
              aria-label={t("dashboard.dailySpendingTitle")}
              onMouseLeave={() => setHover(null)}
            >
              {GRID.map((pos) => (
                <span
                  key={pos}
                  className={`absolute inset-x-0 border-t border-gridline ${pos}`}
                />
              ))}
              <div className="relative flex h-full gap-1">
                {daily.map((amount, i) => (
                  // the whole column is the hover target, so a tiny bar is as easy to hit as a tall one
                  <div
                    key={i}
                    onMouseEnter={() => setHover(i)}
                    className={`flex h-full min-w-0 flex-1 items-end rounded-full ${hover === i ? "bg-surface-2" : ""}`}
                  >
                    {today !== undefined && i + 1 > today ? (
                      <span className="dusk-future block h-1.5 w-full rounded-full" />
                    ) : (
                      <span
                        className={`block w-full rounded-full ${i + 1 === today || hover === i ? "bg-series-2" : "bg-series-1"}`}
                        style={{
                          height: `${(amount / max) * 100}%`,
                          minHeight: amount > 0 ? 4 : 0,
                        }}
                      />
                    )}
                  </div>
                ))}
              </div>
              {hover !== null && (
                <div
                  className="pointer-events-none absolute z-10 -translate-x-1/2 whitespace-nowrap rounded-md border border-border bg-surface-0 px-2.5 py-1.5 text-xs shadow-md"
                  style={{
                    left: `${Math.min(Math.max(((hover + 0.5) / daily.length) * 100, 8), 92)}%`,
                    bottom: `calc(${(daily[hover] / max) * 100}% + 8px)`,
                  }}
                >
                  <p className="text-text-muted">
                    {formatTransactionDate(
                      `${year}-${pad(month)}-${pad(hover + 1)}`,
                    )}
                  </p>
                  <p className="font-semibold tabular-nums text-text-primary">
                    {formatCurrency(daily[hover])}
                  </p>
                </div>
              )}
            </div>
            <div className="mt-1 flex justify-between text-xs tabular-nums text-text-muted">
              <span>1</span>
              <span>{Math.ceil(daily.length / 2)}</span>
              <span>{daily.length}</span>
            </div>
            <dl className="mt-4 grid grid-cols-3 divide-x divide-border border-t border-border pt-4 text-sm">
              {[
                [t("dashboard.dailySpendingTotal"), total],
                [t("dashboard.dailySpendingAverage"), total / elapsed],
                [t("dashboard.dailySpendingPeak"), max],
              ].map(([label, value]) => (
                <div key={label} className="px-4 first:pl-0">
                  <dt className="text-xs text-text-muted">{label}</dt>
                  <dd className="font-display tabular-nums text-text-primary">{formatCurrency(Number(value))}</dd>
                </div>
              ))}
            </dl>
          </>
        )}
      </CardContent>
    </Card>
  );
}
