import { cn } from "../../lib/utils";

export type RiskLevel = "low" | "medium" | "high";

const levels: RiskLevel[] = ["low", "medium", "high"];

export function RiskIndicator({
  level,
  className,
  showLabel = true,
}: {
  level: RiskLevel;
  className?: string;
  showLabel?: boolean;
}) {
  const active = levels.indexOf(level);
  const colors = ["bg-success", "bg-warning", "bg-destructive"];
  return (
    <div className={cn("inline-flex items-center gap-2", className)} role="img" aria-label={`Risk: ${level}`}>
      <div className="flex gap-0.5">
        {levels.map((l, i) => (
          <span
            key={l}
            className={cn(
              "h-1.5 w-4 rounded-full",
              i <= active ? colors[active] : "bg-muted",
            )}
          />
        ))}
      </div>
      {showLabel ? (
        <span className="text-xs font-medium capitalize text-muted-foreground">{level}</span>
      ) : null}
    </div>
  );
}
