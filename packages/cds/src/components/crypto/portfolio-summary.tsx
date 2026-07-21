import { Card, CardContent, CardHeader, CardTitle } from "../../components/ui/card";
import { CdsDonutChart, type DonutSlice } from "../../components/charts/donut-chart";
import { formatCurrency, formatPercent } from "../../lib/utils";
import { cn } from "../../lib/utils";

export interface PortfolioSummaryProps {
  totalValue: number;
  change24h: number;
  allocation: DonutSlice[];
  className?: string;
}

export function PortfolioSummary({
  totalValue,
  change24h,
  allocation,
  className,
}: PortfolioSummaryProps) {
  const positive = change24h >= 0;
  return (
    <Card className={cn(className)}>
      <CardHeader>
        <CardTitle className="text-base">Portfolio</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2">
        <div className="flex flex-col justify-center">
          <p className="text-xs text-muted-foreground">Total value</p>
          <p className="text-3xl font-bold tabular-nums">{formatCurrency(totalValue)}</p>
          <p
            className={cn(
              "mt-1 text-sm font-medium",
              positive ? "text-success" : "text-destructive",
            )}
          >
            {formatPercent(change24h)} (24h)
          </p>
        </div>
        <CdsDonutChart data={allocation} height={160} innerRadius={48} />
      </CardContent>
    </Card>
  );
}
