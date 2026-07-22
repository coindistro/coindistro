import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Earn Admin" };

export default function Page() {
  return (
    <ComingSoon
      title="Earn Admin"
      description="Configure earn products and campaigns."
      module="admin-earn"
      status="Planned"
      expectedFeatures={[
    "Product CRUD",
    "Launchpool setup",
    "Reward analytics",
      ]}
    />
  );
}
