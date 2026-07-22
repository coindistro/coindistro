import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Wallet" };

export default function Page() {
  return (
    <ComingSoon
      title="Wallet"
      description="Deposits, withdrawals, and balances."
      module="wallet"
      status="Planned"
      expectedFeatures={[
    "Multi-asset balances",
    "Deposit addresses",
    "Withdrawal security",
      ]}
    />
  );
}
