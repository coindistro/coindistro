import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "Markets Admin" };

export default function Page() {
  return (
    <PlaceholderPage
      title="Markets Admin"
      description="Pairs, fees, and market configuration."
      module="admin-markets"
    />
  );
}
