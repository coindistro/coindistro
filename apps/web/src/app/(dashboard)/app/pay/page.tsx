import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Pay" };

export default function Page() {
  return (
    <ComingSoon
      title="Pay"
      description="Spend crypto with Coindistro Pay."
      module="pay"
      status="Planned"
      expectedFeatures={[
    "QR payments",
    "Merchant discovery",
    "Transaction history",
      ]}
    />
  );
}
