import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "System Health" };

export default function Page() {
  return (
    <PlaceholderPage
      title="System Health"
      description="Service health, readiness, and dependencies."
      module="admin-health"
    />
  );
}
