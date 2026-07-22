import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Trade" };

export default function Page() {
  return (
    <ComingSoon
      title="Trade"
      description="Spot trading interface."
      module="trade"
      status="Planned"
      expectedFeatures={[
    "Order book",
    "Limit and market orders",
    "Trade history",
      ]}
    />
  );
}
