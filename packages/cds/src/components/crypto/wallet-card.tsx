import { Wallet } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { cn, formatCurrency, formatCrypto } from "@/lib/utils";

export interface WalletBalance {
  asset: string;
  amount: number;
  fiatValue: number;
}

export interface WalletCardProps {
  totalFiat: number;
  balances: WalletBalance[];
  onDeposit?: () => void;
  onWithdraw?: () => void;
  className?: string;
}

export function WalletCard({
  totalFiat,
  balances,
  onDeposit,
  onWithdraw,
  className,
}: WalletCardProps) {
  return (
    <Card className={cn(className)}>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2 text-base">
          <Wallet className="h-4 w-4 text-primary" aria-hidden />
          Wallet
        </CardTitle>
        <div className="flex gap-2">
          {onDeposit ? (
            <Button size="sm" variant="outline" onClick={onDeposit}>
              Deposit
            </Button>
          ) : null}
          {onWithdraw ? (
            <Button size="sm" variant="ghost" onClick={onWithdraw}>
              Withdraw
            </Button>
          ) : null}
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <p className="text-xs text-muted-foreground">Total balance</p>
          <p className="text-2xl font-bold tabular-nums">{formatCurrency(totalFiat)}</p>
        </div>
        <ul className="space-y-2">
          {balances.map((b) => (
            <li
              key={b.asset}
              className="flex items-center justify-between rounded-lg bg-muted/40 px-3 py-2 text-sm"
            >
              <span className="font-medium">{b.asset}</span>
              <span className="text-right">
                <span className="block font-medium tabular-nums">
                  {formatCrypto(b.amount, b.asset)}
                </span>
                <span className="text-xs text-muted-foreground">
                  {formatCurrency(b.fiatValue)}
                </span>
              </span>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}
