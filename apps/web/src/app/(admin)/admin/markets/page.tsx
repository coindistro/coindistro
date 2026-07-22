import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Markets Admin" };

export default function Page() {
  return (
    <ComingSoon
      title="Markets Admin"
      description="Market listings and trading pairs."
      module="admin-markets"
      status="Planned"
      expectedFeatures={[
    "Pair configuration",
    "Market status",
    "Fee schedules",
      ]}
    />
  );
}
