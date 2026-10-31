import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { MonthSelector } from "@/components/layout/MonthSelector";
import { YearSelector } from "@/components/layout/YearSelector";
import { StatCard } from "@/components/dashboard/StatCard";
import { DailySpendingChart } from "@/components/dashboard/DailySpendingChart";
import { SpendingByCategoryCard } from "@/components/dashboard/SpendingByCategoryCard";
import { RecentTransactionsCard } from "@/components/dashboard/RecentTransactionsCard";
import { AlertBanner } from "@/components/insights/AlertBanner";
import { Select } from "@/components/ui/Input";
import { useAccounts } from "@/hooks/useAccounts";
import { useDashboardSummary } from "@/hooks/useDashboard";
import { useTransactionYears } from "@/hooks/useTransactions";
import { formatCurrency, formatSignedCurrency } from "@/lib/format";
import { setConvertToUsd } from "@/lib/convertToUsd";
import { useTranslation } from "@/lib/i18n";

/** Share of income left over after spending (net / real_income). `null` when
 * there was no income to take a share of, rather than a misleading 0%. */
function savingsRate(realIncome: number, net: number): number | null {
  return realIncome > 0 ? (net / realIncome) * 100 : null;
}

function formatPercent(value: number): string {
  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(0)}%`;
}

export function DashboardPage() {
  const { t } = useTranslation();
  const now = new Date();
  // filters live in the URL so a refresh or a shared link keeps them
  const [params, setParams] = useSearchParams();
  const year = Number(params.get("year")) || now.getFullYear();
  const month = Number(params.get("month")) || now.getMonth() + 1;
  // undefined = every account, the default
  const accountId = Number(params.get("account")) || undefined;
  const setFilter = (key: string, value?: number) =>
    setParams(
      (prev) => {
        if (value) prev.set(key, String(value));
        else prev.delete(key);
        return prev;
      },
      { replace: true }
    );
  const { data: years } = useTransactionYears();
  const { data: accounts } = useAccounts();

  // every account is always shown in USD; a single non-USD account can be shown as is
  const account = accounts?.find((a) => a.id === accountId);
  const canConvert = !!account && account.currency !== "USD";
  const [convert, setConvert] = useState(false);
  const converted = !canConvert || convert;
  // before the queries run, since the request header is read when they fire
  setConvertToUsd(converted, account?.currency);
  useEffect(() => () => setConvertToUsd(true), []);

  const { data, isLoading, isError } = useDashboardSummary(year, month, accountId, converted);
  const rate = data ? savingsRate(Number(data.real_income), Number(data.net)) : null;

  return (
    <div className="space-y-6">
      <AlertBanner excludeKeys={["risky_allocation_exceeded"]} />

      <div className="flex flex-wrap items-center gap-3 border-b border-border pb-5">
        <div className="min-w-0 flex-1">
          <MonthSelector month={month} onChange={(m) => setFilter("month", m)} />
        </div>
        <YearSelector years={years ?? [now.getFullYear()]} year={year} onChange={(y) => setFilter("year", y)} />
        <label className="flex items-center gap-2 text-sm text-text-primary has-[:disabled]:text-text-muted">
          <input
            type="checkbox"
            checked={converted}
            disabled={!canConvert}
            className="disabled:cursor-not-allowed"
            onChange={(event) => setConvert(event.target.checked)}
          />
          {t("dashboard.convertToUsd")}
        </label>
        <Select
          aria-label={t("dashboard.accountFilterLabel")}
          value={accountId ?? ""}
          onChange={(event) => setFilter("account", Number(event.target.value) || undefined)}
          className="w-auto max-w-48 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-series-1"
        >
          <option value="">{t("dashboard.allAccounts")}</option>
          {accounts?.map((account) => (
            <option key={account.id} value={account.id}>
              {account.name}
            </option>
          ))}
        </Select>
      </div>

      {isError && (
        <p className="rounded-lg border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger">
          {t("dashboard.errorLoading")}
        </p>
      )}

      <div className="grid grid-cols-2 lg:grid-cols-[1.6fr_1fr_1fr_1fr] lg:divide-x lg:divide-border">
        <StatCard
          hero
          loading={isLoading}
          label={t("dashboard.statNetLabel")}
          value={formatSignedCurrency(data?.net ?? 0)}
          caption={t("dashboard.statNetCaption")}
          tone={Number(data?.net ?? 0) >= 0 ? "success" : "danger"}
        />
        <StatCard
          loading={isLoading}
          label={t("dashboard.statRealIncomeLabel")}
          value={formatCurrency(data?.real_income ?? 0)}
          caption={t("dashboard.statRealIncomeCaption")}
        />
        <StatCard
          loading={isLoading}
          label={t("dashboard.statSpentLabel")}
          value={formatCurrency(data?.spent ?? 0)}
          caption={t("dashboard.statSpentCaption")}
        />
        <StatCard
          loading={isLoading}
          label={t("dashboard.statSavingsRateLabel")}
          value={rate === null ? "—" : formatPercent(rate)}
          caption={t("dashboard.statSavingsRateCaption")}
          tone={rate === null ? "default" : rate >= 0 ? "success" : "danger"}
        />
      </div>

      <div className="grid grid-cols-1 items-start gap-8 lg:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
        <DailySpendingChart
          year={year}
          month={month}
          daily={data?.daily_spending ?? []}
          today={year === now.getFullYear() && month === now.getMonth() + 1 ? now.getDate() : undefined}
          loading={isLoading}
        />
        <SpendingByCategoryCard items={data?.spending_by_category ?? []} loading={isLoading} />
      </div>

      <RecentTransactionsCard year={year} month={month} accountId={accountId} />
    </div>
  );
}
