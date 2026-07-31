"use client";

import { useQuery } from "@tanstack/react-query";
import * as investmentsApi from "@/features/investments/api";

export const exchangeRateQueryKey = ["investments", "exchange-rate"] as const;
export const investmentSettingsQueryKey = ["investments", "settings"] as const;
export const dashboardQueryKey = ["investments", "dashboard"] as const;
export const investmentsListQueryKey = (page: number, perPage: number) =>
  ["investments", "list", page, perPage] as const;
export const investmentQueryKey = (id: string) =>
  ["investments", id] as const;
export const investmentRewardsQueryKey = (id: string) =>
  ["investments", id, "rewards"] as const;
export const rewardHistoryQueryKey = (page: number, perPage: number) =>
  ["investments", "rewards", page, perPage] as const;
export const paymentHistoryQueryKey = (page: number, perPage: number) =>
  ["investments", "payments", page, perPage] as const;
export const withdrawalHistoryQueryKey = (page: number, perPage: number) =>
  ["investments", "withdrawals", page, perPage] as const;
export const notificationsQueryKey = (page: number, perPage: number) =>
  ["investments", "notifications", page, perPage] as const;
export const unreadCountQueryKey = ["investments", "notifications", "unread-count"] as const;

export function useExchangeRate(enabled = true) {
  return useQuery({
    queryKey: exchangeRateQueryKey,
    queryFn: investmentsApi.getExchangeRate,
    enabled,
    staleTime: 60_000,
  });
}

export function useInvestmentSettings(enabled = true) {
  return useQuery({
    queryKey: investmentSettingsQueryKey,
    queryFn: investmentsApi.getInvestmentSettings,
    enabled,
    staleTime: 60_000,
  });
}

export function useDashboard(enabled = true) {
  return useQuery({
    queryKey: dashboardQueryKey,
    queryFn: investmentsApi.getDashboard,
    enabled,
    staleTime: 30_000,
  });
}

export function useInvestmentsList(page = 1, perPage = 20, enabled = true) {
  return useQuery({
    queryKey: investmentsListQueryKey(page, perPage),
    queryFn: () => investmentsApi.listInvestments(page, perPage),
    enabled,
    staleTime: 30_000,
  });
}

export function useInvestment(id: string, enabled = true) {
  return useQuery({
    queryKey: investmentQueryKey(id),
    queryFn: () => investmentsApi.getInvestment(id),
    enabled: enabled && !!id,
    staleTime: 30_000,
  });
}

export function useInvestmentRewards(id: string, enabled = true) {
  return useQuery({
    queryKey: investmentRewardsQueryKey(id),
    queryFn: () => investmentsApi.getInvestmentRewards(id),
    enabled: enabled && !!id,
    staleTime: 30_000,
  });
}

export function useRewardHistory(page = 1, perPage = 20, enabled = true) {
  return useQuery({
    queryKey: rewardHistoryQueryKey(page, perPage),
    queryFn: () => investmentsApi.getRewardHistory(page, perPage),
    enabled,
    staleTime: 30_000,
  });
}

export function usePaymentHistory(page = 1, perPage = 20, enabled = true) {
  return useQuery({
    queryKey: paymentHistoryQueryKey(page, perPage),
    queryFn: () => investmentsApi.getPaymentHistory(page, perPage),
    enabled,
    staleTime: 30_000,
  });
}

export function useWithdrawalHistory(page = 1, perPage = 20, enabled = true) {
  return useQuery({
    queryKey: withdrawalHistoryQueryKey(page, perPage),
    queryFn: () => investmentsApi.getWithdrawalHistory(page, perPage),
    enabled,
    staleTime: 30_000,
  });
}

export function useNotifications(page = 1, perPage = 20, enabled = true) {
  return useQuery({
    queryKey: notificationsQueryKey(page, perPage),
    queryFn: () => investmentsApi.getNotifications(page, perPage),
    enabled,
    staleTime: 15_000,
  });
}

export function useUnreadNotificationCount(enabled = true) {
  return useQuery({
    queryKey: unreadCountQueryKey,
    queryFn: investmentsApi.getUnreadNotificationCount,
    enabled,
    staleTime: 15_000,
    refetchInterval: 30_000,
  });
}
