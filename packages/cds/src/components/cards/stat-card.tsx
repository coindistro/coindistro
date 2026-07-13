import * as React from "react";
import { TrendingDown, TrendingUp } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn, formatPercent } from "@/lib/utils";

export interface StatCardProps {
  title: string;
  value: React.ReactNode;
  description?: string;
  change?: number;
  icon?: React.ReactNode;
  className?: string;
}

export function StatCard({ title, value, description, change, icon, className }: StatCardProps) {
  const positive = change !== undefined && change >= 0;
  return (
    <Card className={cn(className)}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        {icon ? <div className="text-muted-foreground">{icon}</div> : null}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold tracking-tight">{value}</div>
        <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
          {change !== undefined ? (
            <span
              className={cn(
                "inline-flex items-center gap-0.5 font-medium",
                positive ? "text-success" : "text-destructive",
              )}
            >
              {positive ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
              {formatPercent(change)}
            </span>
          ) : null}
          {description ? <span>{description}</span> : null}
        </div>
      </CardContent>
    </Card>
  );
}
