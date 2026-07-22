import { ComingSoon } from "@/features/shared/components/coming-soon";

export const metadata = { title: "AI Bots" };

export default function Page() {
  return (
    <ComingSoon
      title="AI Bots"
      description="Automated trading bots."
      module="ai-bots"
      status="Planned"
      expectedFeatures={[
    "Bot marketplace",
    "Strategy configuration",
    "Performance analytics",
      ]}
    />
  );
}
