import { api } from "@/lib/api/client";
import type {
  InvestmentDashboard,
  InvestmentPlan,
  InvestmentSummary,
  Wallet,
  WalletTransaction,
} from "@/lib/api/types";

/** Fetch user wallet balances (CDT). */
export async function getWallet(): Promise<Wallet> {
  return api.get<Wallet>("/api/v1/wallet");
}

/** Fetch investment dashboard (active/completed + list). */
export async function getInvestments(): Promise<InvestmentDashboard> {
  return api.get<InvestmentDashboard>("/api/v1/earn/investments");
}

/** Fetch a single investment by id. */
export async function getInvestment(id: string): Promise<InvestmentSummary> {
  return api.get<InvestmentSummary>(`/api/v1/earn/investments/${id}`);
}

/** List Genesis investment plans (public). */
export async function getInvestmentPlans(): Promise<InvestmentPlan[]> {
  const data = await api.get<InvestmentPlan[] | { plans?: InvestmentPlan[] }>(
    "/api/v1/earn/plans",
    { auth: false },
  );
  if (Array.isArray(data)) return data;
  if (data && Array.isArray(data.plans)) return data.plans;
  return [];
}

/** Wallet ledger. */
export async function getWalletTransactions(page = 1, perPage = 20) {
  return api.get<{ items?: WalletTransaction[] } | WalletTransaction[]>(
    `/api/v1/wallet/transactions?page=${page}&per_page=${perPage}`,
  );
}
