import { EmptyState } from "@coindistro/cds";
import { WifiOff } from "lucide-react";

export const metadata = { title: "Offline" };

export default function OfflinePage() {
  return (
    <div className="flex min-h-screen items-center justify-center p-6">
      <EmptyState
        icon={<WifiOff className="h-10 w-10" />}
        title="You are offline"
        description="Reconnect to the internet to continue using Coindistro."
      />
    </div>
  );
}
