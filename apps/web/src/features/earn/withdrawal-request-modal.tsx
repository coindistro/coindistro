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
import { Lock, Clock } from "lucide-react";
import {
  calculateWithdrawal,
  formatCurrency,
  getWithdrawalCooldown,
  formatWithdrawalNextAvailable,
  WITHDRAWAL_INTERVAL_DAYS,
  WITHDRAWAL_PROCESSING_HOURS,
} from "@/features/earn/utils";

interface WithdrawalRequestModalProps {
  open: boolean;
  availableBalance: number;
  processingHours: number;
  feePercent: number;
  penaltyPercent: number;
  earlyWithdrawal: boolean;
  isSubmitting: boolean;
  lastWithdrawalAt?: string | null;
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
  lastWithdrawalAt,
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

  // Weekly withdrawal lock (one request every 7 days)
  const cooldown = React.useMemo(
    () => getWithdrawalCooldown(lastWithdrawalAt),
    [lastWithdrawalAt],
  );
  const locked = !cooldown.available;

  const requested = Number(amount);
  const withdrawal = calculateWithdrawal(requested, feePercent, penaltyPercent, earlyWithdrawal);
  const valid =
    !locked &&
    Number.isFinite(requested) &&
    requested > 0 &&
    requested <= availableBalance &&
    withdrawal.net > 0 &&
    confirmed;

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Weekly Withdrawals</DialogTitle>
          <DialogDescription>
            You may submit one withdrawal request every {WITHDRAWAL_INTERVAL_DAYS} days.
            Processing Time: Up to {processingHours || WITHDRAWAL_PROCESSING_HOURS} Hours.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 text-sm">
          {/* Withdrawal status card */}
          <div
            className={`rounded-lg border p-4 space-y-1 ${
              locked
                ? "border-amber-500/30 bg-amber-500/10"
                : "border-emerald-500/30 bg-emerald-500/10"
            }`}
          >
            <div className="flex items-center gap-2">
              <Lock className="h-4 w-4 text-muted-foreground" />
              <p className="font-semibold text-foreground">Next Withdrawal Available</p>
            </div>
            {locked ? (
              <>
                <p className="mt-2 text-lg font-bold text-amber-600 dark:text-amber-400">
                  {cooldown.daysRemaining} Day{cooldown.daysRemaining === 1 ? "" : "s"} Remaining
                </p>
                <p className="mt-1 text-xs text-muted-foreground">
                  Withdrawals are available once every {WITHDRAWAL_INTERVAL_DAYS} days.
                </p>
                {cooldown.nextAvailableAt && (
                  <p className="text-xs font-medium text-foreground">
                    Next withdrawal: {formatWithdrawalNextAvailable(cooldown.nextAvailableAt)}
                  </p>
                )}
              </>
            ) : (
              <>
                <p className="mt-2 text-lg font-bold text-emerald-600 dark:text-emerald-400">
                  Available Now
                </p>
                <p className="mt-1 text-xs text-muted-foreground">
                  You can submit one withdrawal request now.
                </p>
              </>
            )}
          </div>

          {/* Processing time notice */}
          <div className="rounded-lg border border-info/40 bg-info/10 p-4">
            <div className="flex items-center gap-2 text-info">
              <Clock className="h-4 w-4" />
              <p className="font-medium">Withdrawal Processing Time</p>
            </div>
            <p className="mt-1 text-lg font-bold text-foreground">
              {processingHours || WITHDRAWAL_PROCESSING_HOURS} Hours
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Your withdrawal request will be reviewed and processed within{" "}
              {processingHours || WITHDRAWAL_PROCESSING_HOURS} hours.
            </p>
          </div>

          <div className="rounded-lg border bg-muted/30 p-4 space-y-2">
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Available balance</span>
              <strong>{formatCurrency(availableBalance)}</strong>
            </div>
            <div className="flex justify-between gap-3">
              <span className="text-muted-foreground">Estimated arrival</span>
              <strong>Within {processingHours || WITHDRAWAL_PROCESSING_HOURS} hours</strong>
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
              disabled={isSubmitting || locked}
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
              disabled={isSubmitting || locked}
            />
            <span>
              I understand withdrawals take up to {processingHours || WITHDRAWAL_PROCESSING_HOURS}{" "}
              hours and confirm this request.
            </span>
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
            {isSubmitting ? "Submitting..." : locked ? "Locked" : "Confirm withdrawal"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}