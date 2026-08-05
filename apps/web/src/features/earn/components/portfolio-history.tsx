"use client";

export interface PortfolioHistoryEvent {
  id: string;
  title: string;
  amountLabel: string;
  when: string;
  tone?: "positive" | "neutral" | "primary";
}

export function PortfolioHistory({
  events,
}: {
  events: PortfolioHistoryEvent[];
}) {
  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-lg font-bold text-foreground">Investment History</h2>
        <p className="text-sm text-muted-foreground">Portfolio timeline for this position</p>
      </div>
      {events.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-border bg-card/40 p-6 text-center text-sm text-muted-foreground">
          No portfolio events yet. Invest to start your history.
        </div>
      ) : (
        <ol className="relative space-y-0 rounded-[1.25rem] border border-border/60 bg-card/80 p-4">
          {events.map((event, index) => (
            <li key={event.id} className="relative flex gap-4 pb-5 last:pb-0">
              {index < events.length - 1 && (
                <span className="absolute left-[9px] top-5 h-[calc(100%-8px)] w-px bg-border" aria-hidden />
              )}
              <span
                className={`mt-1 h-[18px] w-[18px] shrink-0 rounded-full border-2 ${
                  event.tone === "positive"
                    ? "border-emerald-500 bg-emerald-500/20"
                    : event.tone === "primary"
                      ? "border-primary bg-primary/20"
                      : "border-muted-foreground/40 bg-muted"
                }`}
              />
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <p className="font-semibold text-foreground">{event.title}</p>
                  <p
                    className={`text-sm font-bold tabular-nums ${
                      event.tone === "positive"
                        ? "text-emerald-500"
                        : event.tone === "primary"
                          ? "text-primary"
                          : "text-foreground"
                    }`}
                  >
                    {event.amountLabel}
                  </p>
                </div>
                <p className="mt-0.5 text-xs text-muted-foreground">{event.when}</p>
              </div>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}

export function buildPortfolioHistory(input: {
  capitalUsd: number;
  profitUsd: number;
  portfolioUsd: number;
  capitalNgn: number;
  profitNgn: number;
  startedAt?: string | null;
  hasProfit: boolean;
}): PortfolioHistoryEvent[] {
  const events: PortfolioHistoryEvent[] = [];
  if (input.capitalUsd > 0 || input.capitalNgn > 0) {
    events.push({
      id: "invest",
      title: "Genesis Investment Created",
      amountLabel: `+$${input.capitalUsd.toLocaleString(undefined, { maximumFractionDigits: 2 })}`,
      when: relativeLabel(input.startedAt) || "When invested",
      tone: "neutral",
    });
  }
  if (input.hasProfit && input.profitUsd > 0) {
    events.push({
      id: "profit",
      title: "Genesis Monthly Profit",
      amountLabel: `+$${input.profitUsd.toLocaleString(undefined, { maximumFractionDigits: 2 })}`,
      when: "Today",
      tone: "positive",
    });
    events.push({
      id: "portfolio",
      title: "Portfolio Value Updated",
      amountLabel: `$${input.portfolioUsd.toLocaleString(undefined, { maximumFractionDigits: 2 })}`,
      when: "Today",
      tone: "primary",
    });
  }
  return events;
}

function relativeLabel(iso?: string | null) {
  if (!iso) return "";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "";
  const days = Math.floor((Date.now() - then) / (24 * 60 * 60 * 1000));
  if (days <= 0) return "Today";
  if (days === 1) return "1 day ago";
  if (days < 60) return `${days} days ago`;
  return new Date(iso).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}
