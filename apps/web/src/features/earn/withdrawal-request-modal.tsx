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
import { calculateWithdrawal, formatCurrency } from "@/features/earn/utils";

interface WithdrawalRequestModalProps {
  open: boolean;
  availableBalance: number;
  processingHours: number;
  feePercent: number;
  penaltyPercent: number;
  earlyWithdrawal: boolean;
  isSubmitting: boolean;
  onClose: () => void;
  onConfirm: (amount: number) => Promise<void> | void;
}

export function WithdrawalRequestModal({
  open,
  availableBalance,
  processingHours,
  feePercent,
  penaltyPercent,
  earlyWithdrawal,
  isSubmitting,
  onClose,
  onConfirm,
}: WithdrawalRequestModalProps) {
  const [amount, setAmount] = React.useState("");
  const [confirmed, setConfirmed] = React.useState(false);

  React.useEffect(() => {
    if (!open) return;
    setAmount("");
    setConfirmed(false);
  }, [open]);

  const requested = Number(amount);
  const withdrawal = calculateWithdrawal(requested, feePercent, penaltyPercent, earlyWithdrawal);
  const valid =
    Number.isFinite(requested) &&
    requested > 0 &&
    requested <= availableBalance &&
    withdrawal.net > 0 &&
    confirmed;

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Request withdrawal</DialogTitle>
          <DialogDescription>
            Withdrawals are processed within {processingHours} hours. Review fees before confirming.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 text-sm">
          <div className="rounded-lg border bg-muted/30 p-4 space-y-2">
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Available balance</span>
              <strong>{formatCurrency(availableBalance)}</strong>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Estimated arrival</span>
              <strong>Within {processingHours} hours</strong>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Status</span>
              <strong>Pending review</strong>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="withdraw-amount">Requested amount (NGN)</Label>
            <Input
              id="withdraw-amount"
              type="number"
              min={0}
              max={availableBalance}
              value={amount}
              onChange={(event) => setAmount(event.target.value)}
              disabled={isSubmitting}
            />
          </div>

          {earlyWithdrawal && (
            <p className="rounded-md border border-amber-500/30 bg-amber-500/10 p-3 text-amber-800 dark:text-amber-200">
              Early withdrawal penalty applies because an active investment has not matured.
            </p>
          )}

          <div className="space-y-2 border-t pt-3">
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Processing fee</span>
              <strong>{formatCurrency(withdrawal.fee)}</strong>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Withdrawal fee / penalty</span>
              <strong>{formatCurrency(withdrawal.penalty)}</strong>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Total deductions</span>
              <strong>{formatCurrency(withdrawal.deductions)}</strong>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">You receive</span>
              <strong>{formatCurrency(withdrawal.net)}</strong>
            </div>
          </div>

          <label className="flex items-start gap-2 text-sm">
            <input
              type="checkbox"
              className="mt-1"
              checked={confirmed}
              onChange={(event) => setConfirmed(event.target.checked)}
              disabled={isSubmitting}
            />
            <span>I understand withdrawals take up to 24 hours and confirm this request.</span>
          </label>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button
            type="button"
            disabled={!valid || isSubmitting}
            onClick={() => void onConfirm(requested)}
          >
            {isSubmitting ? "Submitting..." : "Confirm withdrawal"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
