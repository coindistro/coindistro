import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "Workers" };

export default function Page() {
  return (
    <PlaceholderPage
      title="Workers"
      description="Background job pools and queues."
      module="admin-workers"
    />
  );
}
