"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

export interface SidebarNavItem {
  label: string;
  href: string;
  icon?: React.ReactNode;
  badge?: React.ReactNode;
  active?: boolean;
}

export interface SidebarNavProps {
  items: SidebarNavItem[];
  onNavigate?: (href: string) => void;
  className?: string;
  footer?: React.ReactNode;
  header?: React.ReactNode;
}

/** Presentational sidebar navigation (pass Link via onNavigate or wrap labels). */
export function SidebarNav({ items, onNavigate, className, footer, header }: SidebarNavProps) {
  return (
    <nav className={cn("flex h-full flex-col p-4", className)} aria-label="Sidebar">
      {header ? <div className="mb-6">{header}</div> : null}
      <ul className="flex-1 space-y-1">
        {items.map((item) => (
          <li key={item.href}>
            <button
              type="button"
              onClick={() => onNavigate?.(item.href)}
              className={cn(
                "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                item.active
                  ? "bg-sidebar-accent text-primary"
                  : "text-sidebar-foreground hover:bg-muted",
              )}
              aria-current={item.active ? "page" : undefined}
            >
              {item.icon ? <span className="opacity-80">{item.icon}</span> : null}
              <span className="flex-1 text-left">{item.label}</span>
              {item.badge}
            </button>
          </li>
        ))}
      </ul>
      {footer ? <div className="mt-4 border-t border-sidebar-border pt-4">{footer}</div> : null}
    </nav>
  );
}
