import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Signals" };

export default function Page() {
  return (
    <ComingSoon
      title="Signals"
      description="Trading signals and alerts."
      module="signals"
      status="Planned"
      expectedFeatures={[
    "Signal feed",
    "Risk indicators",
    "Subscription tiers",
      ]}
    />
  );
}
