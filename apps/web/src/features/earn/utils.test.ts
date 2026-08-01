import { describe, expect, it } from "vitest";
import {
  buildRewardTimeline,
  calculateInvestment,
  calculateWithdrawal,
  deriveRoiPercent,
  formatCurrency,
  formatRoi,
  getCompletedBusinessDays,
  getProgressPercentage,
  greetingForHour,
} from "./utils";

describe("earn helpers", () => {
  it("calculates the CoinDistro $30 / ₦1,600 / ₦650 investment model", () => {
    const calc = calculateInvestment({
      amountUsd: 30,
      exchangeRate: 1600,
      dailyRewardNgn: 650,
      durationBusinessDays: 20,
      roiPercent: 0,
    });
    expect(calc).toMatchObject({
      amountNgn: 48000,
      dailyEarningsNgn: 650,
      monthlyEarningsNgn: 13000,
      totalEarningsNgn: 13000,
      totalPayoutNgn: 61000,
      businessDaysRemaining: 20,
    });
    expect(calc.roiPercent).toBeCloseTo(27.08, 1);
  });

  it("respects an explicit ROI percent when provided", () => {
    expect(
      calculateInvestment({
        amountUsd: 30,
        exchangeRate: 1600,
        dailyRewardNgn: 650,
        durationBusinessDays: 20,
        roiPercent: 30,
      }).roiPercent,
    ).toBe(30);
  });

  it("calculates withdrawal deductions", () => {
    expect(calculateWithdrawal(10000, 2, 10, true)).toEqual({
      requested: 10000,
      fee: 200,
      penalty: 1000,
      deductions: 1200,
      net: 8800,
    });
  });

  it("formats money, ROI, progress, and greetings", () => {
    expect(formatCurrency(42000)).toBe("₦42,000");
    expect(formatRoi(27.08)).toBe("27.08%");
    expect(deriveRoiPercent(13000, 48000)).toBeCloseTo(27.08, 1);
    expect(getProgressPercentage(10, 20)).toBe(50);
    expect(getProgressPercentage(0, 20)).toBe(100);
    expect(getCompletedBusinessDays(12, 20)).toBe(8);
    expect(greetingForHour(9)).toBe("Good Morning");
    expect(greetingForHour(14)).toBe("Good Afternoon");
    expect(greetingForHour(20)).toBe("Good Evening");
  });

  it("builds a reward timeline", () => {
    expect(buildRewardTimeline(3, 650)).toEqual([
      { day: 1, amount: 650 },
      { day: 2, amount: 650 },
      { day: 3, amount: 650 },
    ]);
  });
});
