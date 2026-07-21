import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "Admin Overview" };

export default function Page() {
  return (
    <PlaceholderPage
      title="Admin Overview"
      description="Control plane for users, products, and platform health."
      module="admin-overview"
    />
  );
}
