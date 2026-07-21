import Link from "next/link";
import { AuthLayout } from "@coindistro/cds";

export default function AuthRouteLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthLayout
      brand={
        <Link href="/" className="flex items-center gap-2">
          <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-sm font-bold text-primary-foreground">
            C
          </span>
          <span className="text-lg font-semibold gradient-text">Coindistro</span>
        </Link>
      }
      footer={
        <p>
          <Link href="/" className="underline-offset-4 hover:underline">
            Back to home
          </Link>
        </p>
      }
    >
      {children}
    </AuthLayout>
  );
}
