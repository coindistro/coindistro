import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "Earn" };

export default function Page() {
  return (
    <ComingSoon
      title="Earn"
      description="Yield products, launchpools, and learn-to-earn."
      module="earn"
      status="Planned"
      expectedFeatures={[
    "Flexible and locked products",
    "Launchpool campaigns",
    "Reward history",
      ]}
    />
  );
}
