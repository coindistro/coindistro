import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Academy Admin" };

export default function Page() {
  return (
    <ComingSoon
      title="Academy Admin"
      description="Manage courses and instructors."
      module="admin-academy"
      status="Planned"
      expectedFeatures={[
    "Course publishing",
    "Enrollment metrics",
    "Content moderation",
      ]}
    />
  );
}
