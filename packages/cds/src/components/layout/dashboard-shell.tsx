"use client";

import * as React from "react";
import { cn } from "../../lib/utils";

export interface DashboardShellProps {
  sidebar?: React.ReactNode;
  topbar?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  contentClassName?: string;
}

/** Reusable authenticated dashboard chrome (sidebar + topbar + content). */
export function DashboardShell({
  sidebar,
  topbar,
  children,
  className,
  contentClassName,
}: DashboardShellProps) {
  return (
    <div className={cn("flex min-h-screen bg-background", className)}>
      {sidebar ? (
        <aside className="hidden w-64 shrink-0 border-r border-sidebar-border bg-sidebar lg:block">
          {sidebar}
        </aside>
      ) : null}
      <div className="flex min-w-0 flex-1 flex-col">
        {topbar ? (
          <header className="sticky top-0 z-sticky border-b bg-background/80 backdrop-blur-md">
            {topbar}
          </header>
        ) : null}
        <main className={cn("flex-1 p-4 sm:p-6 lg:p-8", contentClassName)}>{children}</main>
      </div>
    </div>
  );
}
