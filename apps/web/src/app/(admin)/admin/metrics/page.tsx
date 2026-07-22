import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Metrics" };

export default function Page() {
  return (
    <ComingSoon
      title="Metrics"
      description="Operational and product metrics."
      module="admin-metrics"
      status="Planned"
      expectedFeatures={[
    "Prometheus views",
    "Business KPIs",
    "SLO dashboards",
      ]}
    />
  );
}
