import { useQuery } from "@tanstack/react-query";
import { fetchDashboardSummary } from "@/api/dashboard";

export function useDashboardSummary(year: number, month: number, accountId?: number, converted = true) {
  return useQuery({
    queryKey: ["dashboard-summary", year, month, accountId, converted],
    queryFn: () => fetchDashboardSummary(year, month, accountId),
  });
}
