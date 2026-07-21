import { PlaceholderPage } from "@/features/shared/components/placeholder-page";

export const metadata = { title: "Feature Flags" };

export default function Page() {
  return (
    <PlaceholderPage
      title="Feature Flags"
      description="Runtime feature flag management."
      module="admin-flags"
    />
  );
}
