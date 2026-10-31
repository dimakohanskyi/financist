import { ChevronRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { getCategoryIcon } from "@/lib/icons";
import { formatCurrency } from "@/lib/format";
import { useTranslation } from "@/lib/i18n";
import { translateCategoryName } from "@/lib/categoryLabels";
import type { CategoryBreakdownItem } from "@/types";

interface SpendingByCategoryCardProps {
  items: CategoryBreakdownItem[];
  loading?: boolean;
}

// Chevron, icon and name, percent, amount.
const ROW_GRID =
  "grid grid-cols-[1rem_minmax(0,1fr)_3rem_6.5rem] items-center gap-3";

export function SpendingByCategoryCard({
  items,
  loading,
}: SpendingByCategoryCardProps) {
  const { t } = useTranslation();
  const hasData = items.length > 0;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm font-medium normal-case tracking-normal text-text-secondary">{t("dashboard.spendingByCategoryTitle")}</CardTitle>
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="h-40 motion-safe:animate-pulse rounded bg-surface-2" />
        ) : !hasData ? (
          <p className="py-10 text-center text-sm text-text-muted">
            {t("dashboard.noExpensesThisMonth")}
          </p>
        ) : (
          <>
            <ul className="divide-y divide-gridline">
              {items.map((item) => {
                const Icon = getCategoryIcon(item.icon);
                const children = item.children;
                const row = (
                  <>
                    <span className="flex h-4 w-4 items-center justify-center text-text-muted">
                      {children.length > 0 && (
                        <ChevronRight
                          size={14}
                          className="transition-transform group-open:rotate-90"
                        />
                      )}
                    </span>
                    <span className="flex min-w-0 items-center gap-2">
                      <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-surface-2 text-text-muted">
                        <Icon size={13} />
                      </span>
                      <span className="min-w-0 flex-1">
                        <span className="block truncate text-sm text-text-primary">
                          {translateCategoryName(item.name)}
                        </span>
                        <span className="mt-1.5 block h-1 rounded-full bg-surface-2">
                          <span
                            className="block h-full rounded-full bg-series-1"
                            style={{ width: `${item.percent}%` }}
                          />
                        </span>
                      </span>
                    </span>
                    <span className="text-right text-xs tabular-nums text-text-muted">
                      {item.percent.toFixed(0)}%
                    </span>
                    <span className="text-right text-sm font-medium tabular-nums text-text-primary">
                      {formatCurrency(item.amount)}
                    </span>
                  </>
                );
                return (
                  <li key={item.category_id ?? "other"}>
                    {children.length === 0 ? (
                      <div className={`${ROW_GRID} py-2.5`}>{row}</div>
                    ) : (
                      <details className="group">
                        <summary
                          className={`${ROW_GRID} cursor-pointer list-none py-2.5 [&::-webkit-details-marker]:hidden focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-series-1`}
                        >
                          {row}
                        </summary>
                        <ul className="mb-2 ml-7 space-y-1.5 border-l border-gridline pl-3">
                          {children.map((child) => (
                            <li
                              key={child.category_id}
                              className="flex items-center justify-between gap-3 text-sm"
                            >
                              <span className="flex min-w-0 items-center gap-2 text-text-secondary">
                                <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-text-muted" />
                                <span className="truncate">
                                  {child.category_id === item.category_id
                                    ? t("reports.directSpendLabel")
                                    : translateCategoryName(child.name)}
                                </span>
                              </span>
                              <span className="tabular-nums text-text-primary">
                                {formatCurrency(child.amount)}
                              </span>
                            </li>
                          ))}
                        </ul>
                      </details>
                    )}
                  </li>
                );
              })}
            </ul>
          </>
        )}
      </CardContent>
    </Card>
  );
}
