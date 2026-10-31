import { Link } from "react-router-dom";
import { getCategoryIcon } from "@/lib/icons";
import { formatCurrency, formatTransactionDate } from "@/lib/format";
import { useTranslation } from "@/lib/i18n";
import { categoryPath, translateCategoryName } from "@/lib/categoryLabels";
import { useTransactions } from "@/hooks/useTransactions";
import { useCategories } from "@/hooks/useCategories";

interface RecentTransactionsCardProps {
  year: number;
  month: number;
  accountId?: number;
}

// Date, description, account, category, amount.
const ROW_GRID = "grid grid-cols-[4.5rem_minmax(0,1.6fr)_minmax(0,1fr)_minmax(0,1fr)_6.5rem] items-center gap-3";

export function RecentTransactionsCard({ year, month, accountId }: RecentTransactionsCardProps) {
  const { t } = useTranslation();
  const { data, isLoading } = useTransactions({ year, month, account_id: accountId, page: 1, page_size: 6 });
  const { data: categories } = useCategories();

  return (
    <section className="min-w-0">
      <div className="flex items-baseline justify-between gap-2 border-b border-border pb-2">
        <h2 className="text-sm font-semibold text-text-primary">{t("dashboard.recentTransactionsTitle")}</h2>
        <Link
          to={`/transactions?year=${year}&month=${month}`}
          className="rounded-sm text-xs font-medium text-series-1 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-series-1"
        >
          {t("dashboard.allTransactionsLink")}
        </Link>
      </div>
      {isLoading ? (
        <p className="py-10 text-center text-sm text-text-muted">{t("common.loading")}</p>
      ) : !data?.items.length ? (
        <p className="py-10 text-center text-sm text-text-muted">{t("dashboard.noTransactionsYet")}</p>
      ) : (
        <>
          <div className={`${ROW_GRID} py-2 text-xs text-text-muted`}>
            <span>{t("dashboard.colDate")}</span>
            <span>{t("dashboard.colDescription")}</span>
            <span>{t("dashboard.colAccount")}</span>
            <span>{t("dashboard.colCategory")}</span>
            <span className="text-right">{t("dashboard.colAmount")}</span>
          </div>
          <ul className="divide-y divide-gridline">
            {data.items.map((tx) => {
              const isSplit = tx.splits.length > 0;
              const Icon = getCategoryIcon(tx.category?.icon);
              const isTransfer = tx.type === "transfer";
              const isExpense = tx.type === "expense";
              const categoryLabel = isSplit
                ? tx.splits.map((split) => (split.category ? translateCategoryName(split.category.name) : "?")).join(" + ")
                : tx.category
                  ? categoryPath(tx.category, categories)
                  : null;
              return (
                <li key={tx.id} className={`${ROW_GRID} py-2.5`}>
                  <span className="text-sm tabular-nums text-text-muted">{formatTransactionDate(tx.date)}</span>
                  <span className="flex min-w-0 items-center gap-2">
                    <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-surface-2 text-text-muted">
                      <Icon size={13} />
                    </span>
                    <span className="truncate text-sm text-text-primary">{tx.description}</span>
                  </span>
                  <span className="truncate text-sm text-text-secondary">{tx.account.name}</span>
                  <span className="truncate text-sm text-text-secondary">{categoryLabel ?? ""}</span>
                  <span
                    className={`text-right text-sm font-medium tabular-nums ${
                      isTransfer ? "text-text-muted" : isExpense ? "text-text-primary" : "text-success"
                    }`}
                  >
                    {isTransfer ? "" : isExpense ? "-" : "+"}
                    {formatCurrency(Number(tx.amount), tx.account.currency)}
                  </span>
                </li>
              );
            })}
          </ul>
        </>
      )}
    </section>
  );
}
