import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "P2P" };

export default function Page() {
  return (
    <ComingSoon
      title="P2P"
      description="Peer-to-peer trading marketplace."
      module="p2p"
      status="Planned"
      expectedFeatures={[
    "Buy and sell ads",
    "Escrow flow",
    "Dispute resolution",
      ]}
    />
  );
}
