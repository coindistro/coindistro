"use client";

import { useQuery } from "@tanstack/react-query";
import * as earnApi from "@/features/earn/api";

export const walletQueryKey = ["wallet"] as const;
export const investmentsQueryKey = ["earn", "investments"] as const;
export const investmentPlansQueryKey = ["earn", "plans"] as const;

/** React Query hook for GET /api/v1/wallet */
export function useWallet(enabled = true) {
  return useQuery({
    queryKey: walletQueryKey,
    queryFn: earnApi.getWallet,
    enabled,
    staleTime: 30_000,
  });
}

/** React Query hook for GET /api/v1/earn/investments */
export function useInvestments(enabled = true) {
  return useQuery({
    queryKey: investmentsQueryKey,
    queryFn: earnApi.getInvestments,
    enabled,
    staleTime: 30_000,
  });
}

/** React Query hook for GET /api/v1/earn/plans */
export function useInvestmentPlans(enabled = true) {
  return useQuery({
    queryKey: investmentPlansQueryKey,
    queryFn: earnApi.getInvestmentPlans,
    enabled,
    staleTime: 60_000,
  });
}
