import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Signals Admin" };

export default function Page() {
  return (
    <ComingSoon
      title="Signals Admin"
      description="Publish and moderate trading signals."
      module="admin-signals"
      status="Planned"
      expectedFeatures={[
    "Signal pipeline",
    "Provider management",
    "Quality scores",
      ]}
    />
  );
}
