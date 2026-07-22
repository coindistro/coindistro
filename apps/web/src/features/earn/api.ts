import { api } from "@/lib/api/client";
import type { EarnPortfolio } from "@/lib/api/types";

/** Fetch earn portfolio; returns null when earn is disabled or unavailable. */
export async function getEarnPortfolio(): Promise<EarnPortfolio | null> {
  try {
    return await api.get<EarnPortfolio>("/api/v1/earn/portfolio");
  } catch {
    return null;
  }
}
