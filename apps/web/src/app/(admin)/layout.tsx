import { RequireAuth } from "@/features/authentication/require-auth";
import { AdminChrome } from "@/features/admin/admin-chrome";

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <RequireAuth roles={["admin", "super_admin"]}>
      <AdminChrome>{children}</AdminChrome>
    </RequireAuth>
  );
}
