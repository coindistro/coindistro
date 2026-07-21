import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "Dashboard" };

export default function Page() {
  return (
    <PlaceholderPage
      title="Dashboard"
      description="Your Coindistro overview — balances, activity, and product shortcuts."
      module="dashboard"
    />
  );
}
