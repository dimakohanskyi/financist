import { RefreshCw } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { useFxStatus, useRefreshFxRates } from "@/hooks/useFx";
import { getIntlLocale } from "@/lib/format";
import { useTranslation } from "@/lib/i18n";

export function FxRatesCard() {
  const { t } = useTranslation();
  const { data: status, isLoading } = useFxStatus();
  const refresh = useRefreshFxRates();

  const updated = status?.updated_at
    ? new Intl.DateTimeFormat(getIntlLocale(), { dateStyle: "medium", timeStyle: "short" }).format(new Date(status.updated_at))
    : t("settings.fxNever");

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("settings.fxTitle")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        <p className="text-sm text-text-primary">
          {isLoading ? "…" : status?.rates.map((r) => `1 USD = ${r.rate} ${r.quote}`).join(" · ")}
        </p>
        <p className="text-xs text-text-muted">
          {t("settings.fxSource", { source: status?.source ?? "…" })} · {t("settings.fxUpdated", { date: updated })}
        </p>
        <div className="flex items-center gap-3">
          <Button type="button" variant="ghost" disabled={refresh.isPending} onClick={() => refresh.mutate()}>
            <RefreshCw size={14} className="mr-1.5" />
            {refresh.isPending ? t("settings.fxRefreshing") : t("settings.fxRefresh")}
          </Button>
          {refresh.isError && <span className="text-sm text-danger">{t("settings.fxError")}</span>}
        </div>
        <p className="text-xs text-text-muted">{t("settings.fxHint")}</p>
      </CardContent>
    </Card>
  );
}
