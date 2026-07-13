import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn, formatCurrency, formatPercent } from "@/lib/utils";
import { cdsCryptoAssets } from "@/tokens";

export interface CoinCardProps {
  symbol: string;
  name: string;
  price: number;
  change24h: number;
  iconUrl?: string;
  className?: string;
  onClick?: () => void;
}

export function CoinCard({
  symbol,
  name,
  price,
  change24h,
  iconUrl,
  className,
  onClick,
}: CoinCardProps) {
  const positive = change24h >= 0;
  const brand =
    cdsCryptoAssets[symbol as keyof typeof cdsCryptoAssets] ?? "hsl(var(--cds-primary))";

  return (
    <Card
      className={cn(
        "cursor-pointer transition-shadow hover:shadow-cds-md focus-within:ring-2 focus-within:ring-ring",
        className,
      )}
      onClick={onClick}
      role={onClick ? "button" : undefined}
      tabIndex={onClick ? 0 : undefined}
    >
      <CardContent className="flex items-center gap-3 p-4">
        <div
          className="flex h-10 w-10 items-center justify-center rounded-full text-sm font-bold text-white"
          style={{ background: brand }}
          aria-hidden
        >
          {iconUrl ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={iconUrl} alt="" className="h-10 w-10 rounded-full" />
          ) : (
            symbol.slice(0, 2)
          )}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="font-semibold">{symbol}</span>
            <span className="truncate text-sm text-muted-foreground">{name}</span>
          </div>
          <div className="mt-0.5 text-sm font-medium">{formatCurrency(price)}</div>
        </div>
        <Badge variant={positive ? "success" : "danger"}>{formatPercent(change24h)}</Badge>
      </CardContent>
    </Card>
  );
}
