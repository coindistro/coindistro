import { ArrowDownLeft, ArrowUpRight, RefreshCw } from "lucide-react";
import { cn, formatCrypto } from "../../lib/utils";
import { Badge } from "../../components/ui/badge";

export type TransactionType = "deposit" | "withdrawal" | "transfer" | "trade" | "reward";
export type TransactionStatus = "pending" | "completed" | "failed";

export interface TransactionItemProps {
  type: TransactionType;
  asset: string;
  amount: number;
  status: TransactionStatus;
  timestamp: string | Date;
  counterparty?: string;
  className?: string;
}

const icons = {
  deposit: ArrowDownLeft,
  withdrawal: ArrowUpRight,
  transfer: RefreshCw,
  trade: RefreshCw,
  reward: ArrowDownLeft,
};

const statusVariant = {
  pending: "warning",
  completed: "success",
  failed: "danger",
} as const;

export function TransactionItem({
  type,
  asset,
  amount,
  status,
  timestamp,
  counterparty,
  className,
}: TransactionItemProps) {
  const Icon = icons[type];
  const date = typeof timestamp === "string" ? new Date(timestamp) : timestamp;
  const positive = type === "deposit" || type === "reward";

  return (
    <div
      className={cn(
        "flex items-center gap-3 rounded-lg border bg-card px-3 py-3 transition-colors hover:bg-muted/40",
        className,
      )}
    >
      <div
        className={cn(
          "flex h-9 w-9 items-center justify-center rounded-full",
          positive ? "bg-success/15 text-success" : "bg-muted text-muted-foreground",
        )}
      >
        <Icon className="h-4 w-4" aria-hidden />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium capitalize">{type}</span>
          <Badge variant={statusVariant[status]} className="capitalize">
            {status}
          </Badge>
        </div>
        <p className="truncate text-xs text-muted-foreground">
          {date.toLocaleString()}
          {counterparty ? ` · ${counterparty}` : ""}
        </p>
      </div>
      <div
        className={cn(
          "text-sm font-semibold tabular-nums",
          positive ? "text-success" : "text-foreground",
        )}
      >
        {positive ? "+" : "-"}
        {formatCrypto(Math.abs(amount), asset)}
      </div>
    </div>
  );
}
