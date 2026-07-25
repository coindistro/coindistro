import { RequireAuth } from "@/features/authentication/require-auth";
import { UserDashboardChrome } from "@/features/dashboard/dashboard-chrome";

// Auth-gated user portal — skip static prerender.
export const dynamic = "force-dynamic";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <RequireAuth>
      <UserDashboardChrome>{children}</UserDashboardChrome>
    </RequireAuth>
  );
}
