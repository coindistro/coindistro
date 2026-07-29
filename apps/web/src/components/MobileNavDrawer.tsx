"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import * as Dialog from "@radix-ui/react-dialog";
import { AnimatePresence, motion } from "framer-motion";
import {
  Button,
  Separator,
  cn,
} from "@coindistro/cds";
import {
  LayoutDashboard,
  LogOut,
  Settings,
  Wallet,
  PiggyBank,
  Gift,
  X,
} from "lucide-react";

export type PublicNavLink = {
  name: string;
  href: string;
};

export type AuthNavLink = {
  name: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
};

const DEFAULT_AUTH_LINKS: AuthNavLink[] = [
  { name: "Dashboard", href: "/app/dashboard", icon: LayoutDashboard },
  { name: "Wallet", href: "/app/wallet", icon: Wallet },
  { name: "Earn", href: "/app/earn", icon: PiggyBank },
  { name: "Referrals", href: "/app/referrals", icon: Gift },
  { name: "Settings", href: "/app/settings", icon: Settings },
];

const DRAWER_TRANSITION = { duration: 0.25, ease: [0.32, 0.72, 0, 1] as const };

export interface MobileNavDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  navLinks: PublicNavLink[];
  isAuthenticated: boolean;
  /** True while auth bootstrap is unresolved — hide auth CTAs to avoid flash. */
  authLoading?: boolean;
  onLogout?: () => void | Promise<void>;
  authLinks?: AuthNavLink[];
  drawerId?: string;
}

function isLinkActive(pathname: string, href: string): boolean {
  if (!href || href === "#") return false;
  if (href.startsWith("/#") || href.startsWith("#")) {
    return pathname === "/";
  }
  if (href === "/") return pathname === "/";
  return pathname === href || pathname.startsWith(`${href}/`);
}

/**
 * Full-height mobile navigation drawer (slides in from the right).
 * Radix Dialog provides focus trap, ESC, and ARIA dialog semantics.
 * Framer Motion provides the 250ms slide + fade (CDS-aligned motion).
 */
export function MobileNavDrawer({
  open,
  onOpenChange,
  navLinks,
  isAuthenticated,
  authLoading = false,
  onLogout,
  authLinks = DEFAULT_AUTH_LINKS,
  drawerId = "mobile-nav-drawer",
}: MobileNavDrawerProps) {
  const pathname = usePathname();
  const close = React.useCallback(() => onOpenChange(false), [onOpenChange]);

  const handleLogout = async () => {
    close();
    await onLogout?.();
  };

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open ? (
          <Dialog.Portal forceMount>
            <Dialog.Overlay asChild>
              <motion.div
                className="fixed inset-0 z-[200] bg-black/60"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={DRAWER_TRANSITION}
                aria-hidden
              />
            </Dialog.Overlay>

            <Dialog.Content
              asChild
              onOpenAutoFocus={(e) => {
                e.preventDefault();
                const root = e.currentTarget as HTMLElement;
                const target =
                  root.querySelector<HTMLElement>("[data-drawer-autofocus]") ??
                  root.querySelector<HTMLElement>("a,button");
                target?.focus();
              }}
            >
              <motion.div
                id={drawerId}
                className={cn(
                  "fixed inset-y-0 right-0 z-[300] flex h-[100dvh] w-full max-w-[min(100vw,20rem)] flex-col border-l border-border bg-background shadow-cds-lg outline-none sm:max-w-sm",
                  "overflow-x-hidden",
                )}
                initial={{ x: "100%" }}
                animate={{ x: 0 }}
                exit={{ x: "100%" }}
                transition={DRAWER_TRANSITION}
              >
                <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-4">
                  <Dialog.Title className="text-base font-semibold tracking-tight text-foreground">
                    Menu
                  </Dialog.Title>
                  <Dialog.Description className="sr-only">
                    Site navigation and account actions
                  </Dialog.Description>
                  <Dialog.Close asChild>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      aria-label="Close navigation menu"
                      data-drawer-autofocus
                    >
                      <X className="h-5 w-5" aria-hidden />
                    </Button>
                  </Dialog.Close>
                </div>

                <nav
                  className="min-h-0 flex-1 overflow-y-auto overflow-x-hidden overscroll-contain px-3 py-3"
                  aria-label="Primary"
                >
                  <ul className="flex flex-col gap-0.5">
                    {navLinks.map((link) => {
                      const active = isLinkActive(pathname, link.href);
                      return (
                        <li key={link.name}>
                          <Link
                            href={link.href}
                            onClick={close}
                            aria-current={active ? "page" : undefined}
                            className={cn(
                              "flex w-full items-center rounded-lg px-3 py-3 text-sm font-medium transition-colors duration-cds",
                              active
                                ? "bg-primary/10 text-primary"
                                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                            )}
                          >
                            {link.name}
                          </Link>
                        </li>
                      );
                    })}
                  </ul>
                </nav>

                <div className="mt-auto shrink-0 border-t border-border bg-background px-4 pb-[max(1rem,env(safe-area-inset-bottom))] pt-4">
                  {authLoading ? (
                    <div
                      className="h-24 animate-pulse rounded-lg bg-muted"
                      aria-hidden
                    />
                  ) : isAuthenticated ? (
                    <div className="flex flex-col gap-1">
                      <p className="mb-1 px-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                        Account
                      </p>
                      {authLinks.map((link) => {
                        const Icon = link.icon;
                        const active = isLinkActive(pathname, link.href);
                        return (
                          <Link
                            key={link.href}
                            href={link.href}
                            onClick={close}
                            aria-current={active ? "page" : undefined}
                            className={cn(
                              "flex w-full items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition-colors duration-cds",
                              active
                                ? "bg-primary/10 text-primary"
                                : "text-foreground hover:bg-accent",
                            )}
                          >
                            <Icon className="h-4 w-4 shrink-0 opacity-80" aria-hidden />
                            {link.name}
                          </Link>
                        );
                      })}
                      <Separator className="my-2" />
                      <Button
                        type="button"
                        variant="outline"
                        size="lg"
                        className="w-full justify-center"
                        onClick={() => void handleLogout()}
                      >
                        <LogOut className="h-4 w-4" aria-hidden />
                        Logout
                      </Button>
                    </div>
                  ) : (
                    <div className="flex w-full min-w-0 flex-col gap-3">
                      <Button
                        asChild
                        variant="outline"
                        size="lg"
                        className="w-full min-w-0 justify-center border-border font-semibold"
                      >
                        <Link href="/login" onClick={close}>
                          Login
                        </Link>
                      </Button>
                      <Button
                        asChild
                        variant="primary"
                        size="lg"
                        className="w-full min-w-0 justify-center bg-gradient-to-r from-[#7C3AED] to-[#06B6D4] font-semibold text-white shadow-cds-sm hover:bg-transparent hover:opacity-90"
                      >
                        <Link href="/register" onClick={close}>
                          Sign Up
                        </Link>
                      </Button>
                    </div>
                  )}
                </div>
              </motion.div>
            </Dialog.Content>
          </Dialog.Portal>
        ) : null}
      </AnimatePresence>
    </Dialog.Root>
  );
}
