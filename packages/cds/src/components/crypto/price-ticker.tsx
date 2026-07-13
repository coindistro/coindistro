import { cn, formatCurrency, formatPercent } from "@/lib/utils";

export interface PriceTickerProps {
  symbol: string;
  price: number;
  change24h: number;
  className?: string;
  compact?: boolean;
}

export function PriceTicker({
  symbol,
  price,
  change24h,
  className,
  compact,
}: PriceTickerProps) {
  const positive = change24h >= 0;
  return (
    <div
      className={cn(
        "inline-flex items-center gap-2 rounded-full border bg-card px-3 py-1.5 text-sm shadow-cds-sm",
        className,
      )}
    >
      <span className="font-semibold">{symbol}</span>
      <span className={cn("font-medium", compact && "text-xs")}>{formatCurrency(price)}</span>
      <span className={cn("text-xs font-medium", positive ? "text-success" : "text-destructive")}>
        {formatPercent(change24h)}
      </span>
    </div>
  );
}
