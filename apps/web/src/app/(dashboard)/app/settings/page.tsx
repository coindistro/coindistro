import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Settings" };

export default function Page() {
  return (
    <ComingSoon
      title="Settings"
      description="Preferences, security, and notifications."
      module="settings"
      status="Planned"
      expectedFeatures={[
    "Theme and locale",
    "Notification preferences",
    "Security controls",
      ]}
    />
  );
}
