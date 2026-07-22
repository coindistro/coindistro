import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Settings" };

export default function Page() {
  return (
    <ComingSoon
      title="Settings"
      description="System configuration and policies."
      module="admin-settings"
      status="Planned"
      expectedFeatures={[
    "App settings",
    "Security policies",
    "Environment metadata",
      ]}
    />
  );
}
