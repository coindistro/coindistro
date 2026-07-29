import { RoleGuard } from "@/features/authentication/guards";
import { MERCHANT_ROLES, ADMIN_ROLES } from "@/features/authentication/roles";

export const dynamic = "force-dynamic";

/** Merchant portal — merchants and admins may access. */
export default function MerchantLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <RoleGuard
      roles={[...MERCHANT_ROLES, ...ADMIN_ROLES]}
      fallbackPath="/dashboard"
    >
      {children}
    </RoleGuard>
  );
}
