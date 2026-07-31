import { api } from "@/lib/api/client";
import type {
  ExchangeRate,
  InvestmentSettings,
  EarningsDashboard,
  EarningsInvestment,
  EarningsSummary,
  RewardHistoryItem,
  PaymentHistoryItem,
  WithdrawalHistoryItem,
  InvestmentNotification,
  InitPaymentResponse,
} from "@/features/investments/types";

export async function getExchangeRate(): Promise<ExchangeRate> {
  return api.get<ExchangeRate>("/api/v1/investments/exchange-rate");
}

export async function getInvestmentSettings(): Promise<InvestmentSettings> {
  return api.get<InvestmentSettings>("/api/v1/investments/settings");
}

export async function getDashboard(): Promise<EarningsDashboard> {
  return api.get<EarningsDashboard>("/api/v1/investments/dashboard");
}

export async function initPaystackPayment(amountUsd: number): Promise<InitPaymentResponse> {
  return api.post<InitPaymentResponse>("/api/v1/investments/paystack/init", {
    amount_usd: amountUsd,
    currency: "NGN",
  });
}

export async function initFlutterwavePayment(amountUsd: number): Promise<InitPaymentResponse> {
  return api.post<InitPaymentResponse>("/api/v1/investments/flutterwave/init", {
    amount_usd: amountUsd,
    currency: "NGN",
  });
}

export async function listInvestments(page = 1, perPage = 20): Promise<EarningsSummary[]> {
  const data = await api.get<EarningsSummary[] | { items?: EarningsSummary[] }>(
    `/api/v1/investments/list?page=${page}&per_page=${perPage}`,
  );
  if (Array.isArray(data)) return data;
  return data.items ?? [];
}

export async function getInvestment(id: string): Promise<EarningsInvestment> {
  return api.get<EarningsInvestment>(`/api/v1/investments/${id}`);
}

export async function getInvestmentRewards(id: string): Promise<RewardHistoryItem[]> {
  const data = await api.get<RewardHistoryItem[] | { items?: RewardHistoryItem[] }>(
    `/api/v1/investments/${id}/rewards`,
  );
  if (Array.isArray(data)) return data;
  return data.items ?? [];
}

export async function getRewardHistory(page = 1, perPage = 20): Promise<RewardHistoryItem[]> {
  return api.get<RewardHistoryItem[]>(`/api/v1/investments/rewards?page=${page}&per_page=${perPage}`);
}

export async function getPaymentHistory(page = 1, perPage = 20): Promise<PaymentHistoryItem[]> {
  return api.get<PaymentHistoryItem[]>(`/api/v1/investments/payments?page=${page}&per_page=${perPage}`);
}

export async function getWithdrawalHistory(page = 1, perPage = 20): Promise<WithdrawalHistoryItem[]> {
  return api.get<WithdrawalHistoryItem[]>(`/api/v1/investments/withdrawals?page=${page}&per_page=${perPage}`);
}

export async function requestWithdrawal(investmentId: string | undefined, amountNgn: number): Promise<void> {
  return api.post<void>("/api/v1/investments/withdraw", {
    investment_id: investmentId,
    amount_ngn: amountNgn,
  });
}

export async function getNotifications(page = 1, perPage = 20): Promise<InvestmentNotification[]> {
  return api.get<InvestmentNotification[]>(`/api/v1/investments/notifications?page=${page}&per_page=${perPage}`);
}

export async function markNotificationRead(id: string): Promise<void> {
  return api.put<void>(`/api/v1/investments/notifications/${id}/read`, {});
}

export async function markAllNotificationsRead(): Promise<void> {
  return api.put<void>("/api/v1/investments/notifications/read-all", {});
}

export async function getUnreadNotificationCount(): Promise<{ count: number }> {
  return api.get<{ count: number }>("/api/v1/investments/notifications/unread-count");
}
