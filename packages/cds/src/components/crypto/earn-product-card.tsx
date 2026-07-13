import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";

export interface EarnProductCardProps {
  name: string;
  category: string;
  apr: number;
  risk: "low" | "medium" | "high";
  assets: string[];
  capacityUsedPct?: number;
  featured?: boolean;
  onJoin?: () => void;
  className?: string;
}

const riskVariant = {
  low: "success",
  medium: "warning",
  high: "danger",
} as const;

export function EarnProductCard({
  name,
  category,
  apr,
  risk,
  assets,
  capacityUsedPct = 0,
  featured,
  onJoin,
  className,
}: EarnProductCardProps) {
  return (
    <Card className={cn(featured && "border-primary/40 shadow-cds-glow", className)}>
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div>
            <CardTitle className="text-base">{name}</CardTitle>
            <CardDescription className="capitalize">{category.replace(/_/g, " ")}</CardDescription>
          </div>
          {featured ? <Badge>Featured</Badge> : null}
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-end justify-between">
          <div>
            <p className="text-xs text-muted-foreground">Est. APR</p>
            <p className="text-2xl font-bold tabular-nums text-success">{apr.toFixed(2)}%</p>
          </div>
          <Badge variant={riskVariant[risk]} className="capitalize">
            {risk} risk
          </Badge>
        </div>
        <div className="flex flex-wrap gap-1">
          {assets.map((a) => (
            <Badge key={a} variant="muted">
              {a}
            </Badge>
          ))}
        </div>
        <div className="space-y-1">
          <div className="flex justify-between text-xs text-muted-foreground">
            <span>Capacity</span>
            <span>{capacityUsedPct.toFixed(0)}%</span>
          </div>
          <Progress value={capacityUsedPct} />
        </div>
      </CardContent>
      <CardFooter>
        <Button className="w-full" onClick={onJoin}>
          Join product
        </Button>
      </CardFooter>
    </Card>
  );
}
