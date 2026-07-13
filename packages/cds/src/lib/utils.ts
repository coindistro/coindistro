import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

/** Merge Tailwind classes safely (clsx + tailwind-merge). */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/** Format fiat/crypto amounts for display. */
export function formatCurrency(
  value: number,
  options: { currency?: string; compact?: boolean; maximumFractionDigits?: number } = {},
) {
  const { currency = "USD", compact = false, maximumFractionDigits = 2 } = options;
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    notation: compact ? "compact" : "standard",
    maximumFractionDigits,
  }).format(value);
}

/** Format crypto quantity with up to 8 decimals. */
export function formatCrypto(value: number, symbol = "", digits = 6) {
  const formatted = new Intl.NumberFormat("en-US", {
    maximumFractionDigits: digits,
  }).format(value);
  return symbol ? `${formatted} ${symbol}` : formatted;
}

/** Format percent change with sign. */
export function formatPercent(value: number, digits = 2) {
  const sign = value > 0 ? "+" : "";
  return `${sign}${value.toFixed(digits)}%`;
}

/** Truncate wallet/crypto addresses for UI. */
export function truncateAddress(address: string, start = 6, end = 4) {
  if (!address || address.length <= start + end) return address;
  return `${address.slice(0, start)}…${address.slice(-end)}`;
}
