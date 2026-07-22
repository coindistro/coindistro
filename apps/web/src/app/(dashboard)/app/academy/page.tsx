import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Academy" };

export default function Page() {
  return (
    <ComingSoon
      title="Academy"
      description="Learn crypto with structured courses."
      module="academy"
      status="Planned"
      expectedFeatures={[
    "Course catalog",
    "Progress tracking",
    "Certificates",
      ]}
    />
  );
}
