import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Referrals" };

export default function Page() {
  return (
    <ComingSoon
      title="Referrals"
      description="Platform-wide referral analytics."
      module="admin-referrals"
      status="Planned"
      expectedFeatures={[
    "Network graph",
    "Reward ledger",
    "Fraud controls",
      ]}
    />
  );
}
