import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Invitations" };

export default function Page() {
  return (
    <ComingSoon
      title="Invitations"
      description="Invitation credit administration."
      module="admin-invitations"
      status="Planned"
      expectedFeatures={[
    "Credit grants",
    "Invite audit",
    "Policy controls",
      ]}
    />
  );
}
