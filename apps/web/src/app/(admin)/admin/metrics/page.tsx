import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "Metrics" };

export default function Page() {
  return (
    <PlaceholderPage
      title="Metrics"
      description="Prometheus-backed operational metrics."
      module="admin-metrics"
    />
  );
}
