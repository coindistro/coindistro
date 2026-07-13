import { Sparkles, Crown } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

export function GenesisBadge({
  number,
  className,
}: {
  number?: number;
  className?: string;
}) {
  return (
    <Badge
      className={cn(
        "gap-1 border-primary/30 bg-primary/15 text-primary",
        className,
      )}
      variant="outline"
    >
      <Sparkles className="h-3 w-3" aria-hidden />
      Genesis{number != null ? ` #${number}` : ""}
    </Badge>
  );
}

export function FounderBadge({ className }: { className?: string }) {
  return (
    <Badge
      className={cn(
        "gap-1 border-warning/40 bg-warning/15 text-warning",
        className,
      )}
      variant="outline"
    >
      <Crown className="h-3 w-3" aria-hidden />
      Founder
    </Badge>
  );
}
