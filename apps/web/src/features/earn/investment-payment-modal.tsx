"use client";

import * as React from "react";
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
} from "@coindistro/cds";
import { CreditCard, Sparkles } from "lucide-react";
import { formatCurrency } from "@/features/earn/utils";

interface InvestmentPaymentModalProps {
  open: boolean;
  planName?: string;
  exchangeRate: number;
  minimumAmount: number;
  defaultAmount: number;
  preferredProvider?: "paystack" | "flutterwave";
  roiPercent?: number;
  durationDays?: number;
  dailyRewardNgn?: number;
  isSubmitting: boolean;
  error?: string | null;
  onClose: () => void;
  onConfirm: (provider: "paystack" | "flutterwave", amount: number) => Promise<void> | void;
}

export function InvestmentPaymentModal({
  open,
  planName,
  exchangeRate,
  minimumAmount,
  defaultAmount,
  preferredProvider = "paystack",
  roiPercent,
  durationDays,
  dailyRewardNgn,
  isSubmitting,
  error,
  onClose,
  onConfirm,
}: InvestmentPaymentModalProps) {
  const [provider, setProvider] = React.useState<"paystack" | "flutterwave">(preferredProvider);
  const [amount, setAmount] = React.useState(String(defaultAmount));

  React.useEffect(() => {
    if (!open) return;
    setAmount(String(defaultAmount));
    setProvider(preferredProvider);
  }, [defaultAmount, open, preferredProvider]);

  const PAYSTACK_PAYMENT_LINK = "https://paystack.shop/pay/coindistro";

  const handlePaystackClick = React.useCallback(() => {
    // TODO:
    // Restore API-based Paystack initialization
    // after payment gateway integration is completed.
    // Current implementation uses hosted Paystack Payment Link.
    const popup = window.open(PAYSTACK_PAYMENT_LINK, "_blank");
    if (!popup) {
      window.location.href = PAYSTACK_PAYMENT_LINK;
    }
  }, []);

  const resolvedAmount = Number(amount);
  const equivalentNgn = Number.isFinite(resolvedAmount) ? resolvedAmount * exchangeRate : 0;
  const valid = Number.isFinite(resolvedAmount) && resolvedAmount >= minimumAmount;

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <div className="flex items-center gap-2 text-primary">
            <CreditCard className="h-5 w-5" />
            <DialogTitle>{planName ? `Invest in ${planName}` : "Confirm investment"}</DialogTitle>
          </div>
          <DialogDescription>
            Choose Paystack or Flutterwave. We create the investment, redirect you to checkout, then activate it after verification.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="provider">Payment provider</Label>
            <div className="grid grid-cols-2 gap-2">
              <Button
                type="button"
                variant={provider === "paystack" ? "primary" : "outline"}
                onClick={() => setProvider("paystack")}
                disabled
              >
                Paystack
              </Button>
              <Button
                type="button"
                variant={provider === "flutterwave" ? "primary" : "outline"}
                onClick={() => setProvider("flutterwave")}
                disabled
              >
                Flutterwave
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">Temporarily using hosted Paystack payment link.</p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="amount">Investment amount (USD)</Label>
            <Input
              id="amount"
              type="number"
              min={minimumAmount}
              step="1"
              value={amount}
              onChange={(event) => setAmount(event.target.value)}
              disabled={isSubmitting}
            />
            <p className="text-xs text-muted-foreground">Minimum investment is ${minimumAmount}.</p>
          </div>

          <div className="rounded-lg border bg-muted/30 p-4 text-sm space-y-2">
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">USD amount</span>
              <span className="font-medium">${Number.isFinite(resolvedAmount) ? resolvedAmount.toLocaleString() : "—"}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">NGN equivalent</span>
              <span className="font-medium">{formatCurrency(equivalentNgn)}</span>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Exchange rate</span>
              <span className="font-medium">1 USD = {formatCurrency(exchangeRate)}</span>
            </div>
            {roiPercent != null && (
              <div className="flex justify-between gap-3">
                <span className="text-muted-foreground">ROI</span>
                <span className="font-medium">{roiPercent}%</span>
              </div>
            )}
            {durationDays != null && (
              <div className="flex justify-between gap-3">
                <span className="text-muted-foreground">Duration</span>
                <span className="font-medium">{durationDays} business days</span>
              </div>
            )}
            {dailyRewardNgn != null && (
              <div className="flex justify-between gap-3">
                <span className="text-muted-foreground">Daily payout</span>
                <span className="font-medium">{formatCurrency(dailyRewardNgn)}</span>
              </div>
            )}
          </div>

          <div className="rounded-lg border border-primary/20 bg-primary/5 p-4 text-sm text-muted-foreground">
            <div className="flex items-center gap-2 text-foreground">
              <Sparkles className="h-4 w-4" />
              <span className="font-medium">Secure checkout</span>
            </div>
            <p className="mt-1">
              You will be redirected to Paystack to complete payment.
            </p>
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

          {error ? (
            <div
              role="alert"
              className="rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive"
            >
              {error}
            </div>
          ) : null}
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button
            type="button"
            onClick={handlePaystackClick}
            disabled={!valid}
          >
            Pay with Paystack
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
