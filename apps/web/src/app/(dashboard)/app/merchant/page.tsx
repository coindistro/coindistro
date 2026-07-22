import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Merchant" };

export default function Page() {
  return (
    <ComingSoon
      title="Merchant"
      description="Accept crypto payments as a merchant."
      module="merchant"
      status="Planned"
      expectedFeatures={[
    "Payment links",
    "Settlement",
    "Merchant dashboard",
      ]}
    />
  );
}
