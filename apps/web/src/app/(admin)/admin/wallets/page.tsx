import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Wallets Admin" };

export default function Page() {
  return (
    <ComingSoon
      title="Wallets Admin"
      description="Wallet operations and freezes."
      module="admin-wallets"
      status="Planned"
      expectedFeatures={[
    "Balance overview",
    "Freeze/unfreeze",
    "Hot wallet health",
      ]}
    />
  );
}
