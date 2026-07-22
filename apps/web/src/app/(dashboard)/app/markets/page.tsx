import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Markets" };

export default function Page() {
  return (
    <ComingSoon
      title="Markets"
      description="Live prices, charts, and market discovery."
      module="markets"
      status="Planned"
      expectedFeatures={[
    "Spot market listings",
    "Price charts",
    "Watchlists",
      ]}
    />
  );
}
