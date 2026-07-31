import { describe, expect, it } from "vitest";
import { buildRewardTimeline, calculateInvestment, calculateWithdrawal, formatCurrency, getProgressPercentage } from "./utils";

describe("earn helpers", () => {
  it("calculates investment projections from configuration", () => {
    expect(calculateInvestment({ amountUsd: 30, exchangeRate: 1600, dailyRewardNgn: 500, durationBusinessDays: 20, roiPercent: 33.33 })).toMatchObject({ amountNgn: 48000, monthlyEarningsNgn: 10000, totalEarningsNgn: 10000, totalPayoutNgn: 58000 });
  });
  it("calculates withdrawal deductions", () => {
    expect(calculateWithdrawal(10000, 2, 10, true)).toEqual({ requested: 10000, fee: 200, penalty: 1000, deductions: 1200, net: 8800 });
  });
  it("formats money and progress", () => { expect(formatCurrency(42000)).toBe("₦42,000"); expect(getProgressPercentage(10, 20)).toBe(50); expect(getProgressPercentage(0, 20)).toBe(100); });
  it("builds a reward timeline", () => { expect(buildRewardTimeline(3, 650)).toHaveLength(3); });
});
