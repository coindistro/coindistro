import { RoleGuard } from "@/features/authentication/guards";
import { AdminChrome } from "@/features/admin/admin-chrome";
import { ADMIN_ROLES } from "@/features/authentication/roles";

// Auth-gated control plane — skip static prerender (Radix Slot + client session).
export const dynamic = "force-dynamic";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <RoleGuard roles={[...ADMIN_ROLES]} fallbackPath="/dashboard">
      <AdminChrome>{children}</AdminChrome>
    </RoleGuard>
  );
}
