import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Merchants" };

export default function Page() {
  return (
    <ComingSoon
      title="Merchants"
      description="Merchant onboarding and approvals."
      module="admin-merchants"
      status="Planned"
      expectedFeatures={[
    "KYB review",
    "Settlement settings",
    "Risk scoring",
      ]}
    />
  );
}
