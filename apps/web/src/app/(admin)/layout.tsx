import { RequireAuth } from "@/features/authentication/require-auth";
import { AdminChrome } from "@/features/admin/admin-chrome";
import { ADMIN_ROLES } from "@/lib/api/types";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <RequireAuth roles={[...ADMIN_ROLES]}>
      <AdminChrome>{children}</AdminChrome>
    </RequireAuth>
  );
}
