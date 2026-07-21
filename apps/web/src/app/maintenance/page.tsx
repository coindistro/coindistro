import { EmptyState } from "@coindistro/cds";
import { Wrench } from "lucide-react";

export const metadata = { title: "Maintenance" };

export default function MaintenancePage() {
  return (
    <div className="flex min-h-screen items-center justify-center p-6">
      <EmptyState
        icon={<Wrench className="h-10 w-10" />}
        title="Scheduled maintenance"
        description="Coindistro is temporarily unavailable. Please check back shortly."
      />
    </div>
  );
}
