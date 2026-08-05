"use client";

import { useState, useCallback, useEffect } from "react";
import { Button, Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@coindistro/cds";
import { CreditCard, Sparkles } from "lucide-react";
import {
  ENABLED_PLANS,
  PAYSTACK_PAYMENT_LINK,
  type InvestmentPlanConfig,
} from "@/features/earn/config/investment-plans";
import { formatCurrency } from "@/features/earn/utils";

interface InvestmentModalProps {
  open: boolean;
  plan: InvestmentPlanConfig | null;
  exchangeRate: number;
  onClose: () => void;
}

export function InvestmentModal({ open, plan, exchangeRate, onClose }: InvestmentModalProps) {
  const [paying, setPaying] = useState(false);
  const [provider, setProvider] = useState<"paystack" | "flutterwave">("paystack");

  // Genesis is the first/default investment option when no plan is pre-selected.
  const resolvedPlan = plan ?? ENABLED_PLANS[0] ?? null;

  useEffect(() => {
    if (!open) {
      setPaying(false);
      setProvider("paystack");
    }
  }, [open]);

  const handlePayClick = useCallback(() => {
    // TODO:
    // Restore API-based Paystack / Flutterwave initialization
    // after payment gateway integration is completed.
    // Current implementation uses hosted Paystack Payment Link.
    setPaying(true);
    const popup = window.open(PAYSTACK_PAYMENT_LINK, "_blank");
    if (!popup) {
      window.location.href = PAYSTACK_PAYMENT_LINK;
    }
    setPaying(false);
  }, []);

  if (!resolvedPlan) return null;

  const ngnValue = resolvedPlan.usdAmount * exchangeRate;

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <div className="flex items-center gap-2 text-primary">
            <CreditCard className="h-5 w-5" />
            <DialogTitle>{resolvedPlan.name} Plan</DialogTitle>
          </div>
          <DialogDescription>{resolvedPlan.description}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="rounded-lg border bg-muted/30 p-4 text-sm space-y-2">
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Investment</span>
              <span className="font-medium">${resolvedPlan.usdAmount.toLocaleString()}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Exchange Rate</span>
              <span className="font-medium">1 USD = {formatCurrency(exchangeRate)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Equivalent</span>
              <span className="font-medium text-fuchsia-600">{formatCurrency(ngnValue)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Expected ROI</span>
              <span className="font-medium text-primary">{resolvedPlan.roiPercent}%</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Daily Reward</span>
              <span className="font-medium text-amber-600">{formatCurrency(resolvedPlan.dailyRewardNgn)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Duration</span>
              <span className="font-medium">{resolvedPlan.workingDays} Business Days</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Expected Payout</span>
              <span className="font-medium text-emerald-600">
                {formatCurrency(ngnValue + resolvedPlan.totalReturnNgn)}
              </span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Total Return</span>
              <span className="font-medium text-emerald-600">{formatCurrency(resolvedPlan.totalReturnNgn)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Referral Reward</span>
              <span className="font-medium text-cyan-600">{resolvedPlan.referralBonusPercent}%</span>
            </div>
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium text-foreground">Payment Methods</p>
            <div className="grid grid-cols-2 gap-2">
              <Button
                type="button"
                variant={provider === "paystack" ? "primary" : "outline"}
                onClick={() => setProvider("paystack")}
                disabled={paying}
              >
                Paystack
              </Button>
              <Button
                type="button"
                variant={provider === "flutterwave" ? "primary" : "outline"}
                onClick={() => setProvider("flutterwave")}
                disabled={paying}
              >
                Flutterwave
              </Button>
            </div>
            {provider === "flutterwave" ? (
              <p className="text-xs text-muted-foreground">
                Flutterwave checkout uses the same temporary hosted payment flow until full gateway wiring is restored.
              </p>
            ) : (
              <p className="text-xs text-muted-foreground">
                You will be redirected to Paystack to complete payment.
              </p>
            )}
          </div>

          <div className="rounded-lg border border-primary/20 bg-primary/5 p-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-2 text-foreground">
              <Sparkles className="h-4 w-4" />
              <span className="font-medium">Secure checkout</span>
            </div>
            <p className="mt-1">
              Processing Time: Up to 24 Hours after payment confirmation for wallet credit.
            </p>
          </div>

          <div className="rounded-lg border border-info/40 bg-info/10 p-4 text-sm text-info">
            <p className="font-medium">Payment Processing Notice</p>
            <p className="mt-1">
              After completing your payment, your CoinDistro wallet will be credited manually within{" "}
              <strong>24 hours</strong> after payment confirmation.
            </p>
            <p className="mt-1">
              Please ensure you use the same email address associated with your CoinDistro account when making payment.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose} disabled={paying}>
            Cancel
          </Button>
          <Button type="button" onClick={handlePayClick} disabled={paying}>
            {paying
              ? "Redirecting..."
              : provider === "flutterwave"
                ? "Pay with Flutterwave"
                : "Pay with Paystack"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
