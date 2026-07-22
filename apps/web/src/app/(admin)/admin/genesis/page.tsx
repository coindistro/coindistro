import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Genesis Members" };

export default function Page() {
  return (
    <ComingSoon
      title="Genesis Members"
      description="Manage the Genesis membership program."
      module="admin-genesis"
      status="Planned"
      expectedFeatures={[
    "Member directory",
    "Slot allocation",
    "Genesis analytics",
      ]}
    />
  );
}
