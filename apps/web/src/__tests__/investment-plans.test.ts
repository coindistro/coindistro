import { describe, it, expect } from "vitest";
import {
  INVESTMENT_PLANS,
  ENABLED_PLANS,
  ROI_LADDER,
  DEFAULT_EXCHANGE_RATE,
  WITHDRAWAL_INTERVAL_DAYS,
  WITHDRAWAL_PROCESSING_HOURS,
  deriveDailyRewardNgn,
  deriveTotalReturnNgn,
} from "@/features/earn/config/investment-plans";
import {
  getWithdrawalCooldown,
  formatWithdrawalNextAvailable,
  WITHDRAWAL_INTERVAL_DAYS as UTIL_INTERVAL,
  WITHDRAWAL_PROCESSING_HOURS as UTIL_HOURS,
} from "@/features/earn/utils";

describe("Investment Plans — Genesis tier", () => {
  it("includes a Genesis plan as the smallest tier", () => {
    const genesis = INVESTMENT_PLANS.find((p) => p.name === "Genesis");
    expect(genesis).toBeDefined();
    expect(genesis?.usdAmount).toBe(10);
  });

  it("Genesis has 18% ROI", () => {
    const genesis = INVESTMENT_PLANS.find((p) => p.name === "Genesis");
    expect(genesis?.roiPercent).toBe(18);
  });

  it("Genesis is the first enabled plan by displayOrder", () => {
    const sorted = [...ENABLED_PLANS].sort((a, b) => a.displayOrder - b.displayOrder);
    expect(sorted[0].name).toBe("Genesis");
  });
});

describe("ROI Ladder", () => {
  it("starts at 18% and increases by exactly 3% per tier", () => {
    expect(ROI_LADDER[0].roiPercent).toBe(18);
    for (let i = 1; i < ROI_LADDER.length; i++) {
      const diff = ROI_LADDER[i].roiPercent - ROI_LADDER[i - 1].roiPercent;
      expect(diff).toBe(3);
    }
  });

  it("has exactly 6 tiers (Genesis → Enterprise)", () => {
    expect(ROI_LADDER).toHaveLength(6);
    expect(ROI_LADDER[5].tier).toBe("Enterprise");
    expect(ROI_LADDER[5].roiPercent).toBe(33);
  });

  it("each plan's roiPercent matches the ladder", () => {
    for (const plan of INVESTMENT_PLANS) {
      const ladderEntry = ROI_LADDER.find((l) => l.tier === plan.name);
      expect(ladderEntry).toBeDefined();
      expect(plan.roiPercent).toBe(ladderEntry!.roiPercent);
    }
  });
});

describe("Investment Calculator — Genesis", () => {
  it("derives correct daily reward for Genesis ($10, 18%, ₦1400)", () => {
    const daily = deriveDailyRewardNgn(10, 18, 1400, 20);
    // capital = 10 * 1400 = 14000; total return = 14000 * 0.18 = 2520; daily = 2520 / 20 = 126
    expect(daily).toBe(126);
  });

  it("derives correct total return for Genesis ($10, 18%, ₦1400)", () => {
    const total = deriveTotalReturnNgn(10, 18, 1400);
    expect(total).toBe(2520);
  });

  it("derives correct total return for Enterprise ($1000, 33%, ₦1400)", () => {
    const total = deriveTotalReturnNgn(1000, 33, 1400);
    // 1000 * 1400 = 1,400,000; * 0.33 = 462,000
    expect(total).toBe(462000);
  });
});

describe("Weekly Withdrawal Lock", () => {
  it("exports 7-day interval and 24-hour processing", () => {
    expect(WITHDRAWAL_INTERVAL_DAYS).toBe(7);
    expect(WITHDRAWAL_PROCESSING_HOURS).toBe(24);
    expect(UTIL_INTERVAL).toBe(7);
    expect(UTIL_HOURS).toBe(24);
  });

  it("returns available when no prior withdrawal", () => {
    const cooldown = getWithdrawalCooldown(null);
    expect(cooldown.available).toBe(true);
    expect(cooldown.daysRemaining).toBe(0);
  });

  it("returns locked when last withdrawal was < 7 days ago", () => {
    const recent = new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString();
    const cooldown = getWithdrawalCooldown(recent);
    expect(cooldown.available).toBe(false);
    expect(cooldown.daysRemaining).toBeGreaterThan(0);
    expect(cooldown.daysRemaining).toBeLessThanOrEqual(7);
  });

  it("returns available when last withdrawal was > 7 days ago", () => {
    const old = new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString();
    const cooldown = getWithdrawalCooldown(old);
    expect(cooldown.available).toBe(true);
    expect(cooldown.daysRemaining).toBe(0);
  });

  it("formats the next available date", () => {
    const iso = "2026-08-14T10:00:00.000Z";
    const formatted = formatWithdrawalNextAvailable(iso);
    expect(formatted).toMatch(/August/);
    expect(formatted).toMatch(/14/);
  });
});

describe("Default Exchange Rate", () => {
  it("defaults to ₦1,400 / $1", () => {
    expect(DEFAULT_EXCHANGE_RATE).toBe(1400);
  });
});