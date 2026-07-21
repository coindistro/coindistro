import { RequireAuth } from "@/features/authentication/require-auth";
import { UserDashboardChrome } from "@/features/dashboard/dashboard-chrome";

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
