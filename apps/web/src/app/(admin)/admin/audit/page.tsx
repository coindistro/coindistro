import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Audit Logs" };

export default function Page() {
  return (
    <ComingSoon
      title="Audit Logs"
      description="Immutable audit trail for sensitive actions."
      module="admin-audit"
      status="Planned"
      expectedFeatures={[
    "Search and filter",
    "Export",
    "Integrity checks",
      ]}
    />
  );
}
