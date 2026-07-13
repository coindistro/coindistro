import * as React from "react";
import { cn } from "@/lib/utils";

export interface AuthLayoutProps {
  children: React.ReactNode;
  brand?: React.ReactNode;
  footer?: React.ReactNode;
  className?: string;
}

/** Centered auth shell for login / register / reset flows. */
export function AuthLayout({ children, brand, footer, className }: AuthLayoutProps) {
  return (
    <div className={cn("flex min-h-screen flex-col items-center justify-center bg-background px-4", className)}>
      <div className="mb-8">{brand}</div>
      <div className="w-full max-w-md rounded-xl border bg-card p-6 shadow-cds-md sm:p-8">
        {children}
      </div>
      {footer ? <div className="mt-6 text-center text-sm text-muted-foreground">{footer}</div> : null}
    </div>
  );
}
