import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "Wallet" };

export default function Page() {
  return (
    <PlaceholderPage
      title="Wallet"
      description="Balances, deposits, withdrawals, and ledger history."
      module="wallet"
    />
  );
}
