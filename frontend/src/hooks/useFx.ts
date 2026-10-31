import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchFxStatus, refreshFxRates } from "@/api/fx";

export function useFxStatus() {
  return useQuery({ queryKey: ["fx"], queryFn: fetchFxStatus });
}

export function useRefreshFxRates() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: refreshFxRates,
    // new rates change every converted amount
    onSuccess: () => queryClient.invalidateQueries(),
  });
}
