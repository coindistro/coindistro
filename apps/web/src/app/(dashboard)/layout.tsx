import { AuthGuard } from "@/features/authentication/guards";
import { UserDashboardChrome } from "@/features/dashboard/dashboard-chrome";

// Auth-gated user portal — skip static prerender.
export const dynamic = "force-dynamic";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthGuard>
      <UserDashboardChrome>{children}</UserDashboardChrome>
    </AuthGuard>
  );
}
