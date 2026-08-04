"use client";

import { useState, useCallback, useEffect } from "react";
import { Button, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@coindistro/cds";
import { CreditCard, Sparkles } from "lucide-react";
import type { InvestmentPlanConfig } from "@/features/earn/config/investment-plans";
import { PAYSTACK_PAYMENT_LINK } from "@/features/earn/config/investment-plans";
import { formatCurrency } from "@/features/earn/utils";

interface InvestmentModalProps {
  open: boolean;
  plan: InvestmentPlanConfig | null;
  exchangeRate: number;
  onClose: () => void;
}

export function InvestmentModal({ open, plan, exchangeRate, onClose }: InvestmentModalProps) {
  const [paying, setPaying] = useState(false);

  useEffect(() => {
    if (!open) {
      setPaying(false);
    }
  }, [open]);

  const handlePaystackClick = useCallback(() => {
    // TODO:
    // Restore API-based Paystack initialization
    // after payment gateway integration is completed.
    // Current implementation uses hosted Paystack Payment Link.
    const popup = window.open(PAYSTACK_PAYMENT_LINK, "_blank");
    if (!popup) {
      window.location.href = PAYSTACK_PAYMENT_LINK;
    }
  }, []);

  if (!plan) return null;

  const ngnValue = plan.usdAmount * exchangeRate;

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <div className="flex items-center gap-2 text-primary">
            <CreditCard className="h-5 w-5" />
            <DialogTitle>{plan.name}</DialogTitle>
          </div>
          <DialogDescription>{plan.description}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="rounded-lg border bg-muted/30 p-4 text-sm space-y-2">
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Investment</span>
              <span className="font-medium">${plan.usdAmount.toLocaleString()}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Exchange Rate</span>
              <span className="font-medium">1 USD = {formatCurrency(exchangeRate)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Investment Value</span>
              <span className="font-medium text-fuchsia-600">{formatCurrency(ngnValue)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Daily Reward</span>
              <span className="font-medium text-amber-600">{formatCurrency(plan.dailyRewardNgn)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Working Days</span>
              <span className="font-medium">{plan.workingDays}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Monthly Reward</span>
              <span className="font-medium text-emerald-600">{formatCurrency(plan.monthlyRewardNgn)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Referral Reward</span>
              <span className="font-medium text-cyan-600">{plan.referralBonusPercent}%</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Minimum Referrals</span>
              <span className="font-medium">{plan.minReferrals}</span>
            </div>
          </div>

          <div className="rounded-lg border border-primary/20 bg-primary/5 p-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-2 text-foreground">
              <Sparkles className="h-4 w-4" />
              <span className="font-medium">Secure checkout</span>
            </div>
            <p className="mt-1">You will be redirected to Paystack to complete payment.</p>
          </div>

          <div className="rounded-lg border border-info/40 bg-info/10 p-4 text-sm text-info">
            <p className="font-medium">Payment Processing Notice</p>
            <p className="mt-1">
              After completing your payment, your CoinDistro wallet will be credited manually within <strong>24 hours</strong> after payment confirmation.
            </p>
            <p className="mt-1">
              Please ensure you use the same email address associated with your CoinDistro account when making payment.
            </p>
            <p className="mt-1">
              This is a temporary payment process while automated wallet funding is being finalized.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose} disabled={paying}>
            Cancel
          </Button>
          <Button type="button" onClick={handlePaystackClick} disabled={paying}>
            Pay with Paystack
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}