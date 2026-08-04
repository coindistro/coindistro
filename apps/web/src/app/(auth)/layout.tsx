import Image from "next/image";
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
        <Link href="/" className="flex flex-col items-center gap-3">
          <div className="relative h-14 w-14 animate-cds-fade-in md:h-16 md:w-16 lg:h-[72px] lg:w-[72px]">
            <Image
              src="/coindistro-logo.png"
              alt="Coindistro"
              fill
              className="object-contain"
              priority
            />
          </div>
          <span className="text-xl font-bold gradient-text">Coindistro</span>
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
