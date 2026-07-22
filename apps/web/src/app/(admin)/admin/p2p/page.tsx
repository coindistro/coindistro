import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "P2P Admin" };

export default function Page() {
  return (
    <ComingSoon
      title="P2P Admin"
      description="Moderate P2P marketplace activity."
      module="admin-p2p"
      status="Planned"
      expectedFeatures={[
    "Ad moderation",
    "Dispute queue",
    "Merchant risk",
      ]}
    />
  );
}
