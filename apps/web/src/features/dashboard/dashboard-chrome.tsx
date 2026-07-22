"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Avatar,
  AvatarFallback,
  Button,
  DashboardShell,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  SidebarNav,
  Topbar,
  useTheme,
} from "@coindistro/cds";
import { Bell, Command, LogOut, Moon, Search, Shield, Sun, User } from "lucide-react";
import { useAuth } from "@/features/authentication/auth-provider";
import { userNavItems } from "@/features/dashboard/nav";
import { Breadcrumbs } from "@/features/shared/components/breadcrumbs";
import { useCommandPalette } from "@/features/shared/providers/command-palette-provider";

export function UserDashboardChrome({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout, isAdmin } = useAuth();
  const { theme, setTheme } = useTheme();
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const [search, setSearch] = React.useState("");
  const { setOpen: setCommandOpen } = useCommandPalette();

  const items = userNavItems.map((item) => ({
    label: item.label,
    href: item.href,
    icon: <item.icon className="h-4 w-4" />,
    active: pathname === item.href || pathname.startsWith(item.href + "/"),
  }));

  const brand = (
    <Link href="/app/dashboard" className="flex items-center gap-2 px-1">
      <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-sm font-bold text-primary-foreground">
        C
      </span>
      <span className="font-semibold">Coindistro</span>
    </Link>
  );

  const sidebar = (
    <SidebarNav
      header={brand}
      items={items}
      onNavigate={(href) => {
        setMobileOpen(false);
        router.push(href);
      }}
      footer={
        <p className="px-3 text-xs text-muted-foreground">
          {user?.email ?? "Signed in"}
        </p>
      }
    />
  );

  const topbar = (
    <Topbar
      onMenuClick={() => setMobileOpen((o) => !o)}
      search={
        <div className="relative">
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search modules…"
            aria-label="Search"
            className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-2 pl-9 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          />
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" aria-hidden />
        </div>
      }
      actions={
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Search (Cmd+K)"
            onClick={() => setCommandOpen(true)}
            className="hidden sm:inline-flex"
          >
            <Command className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Notifications"
            onClick={() => router.push("/app/notifications")}
          >
            <Bell className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            aria-label="Toggle theme"
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
          >
            {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon-sm" aria-label="User menu">
                <Avatar className="h-7 w-7">
                  <AvatarFallback className="text-xs">
                    {(user?.display_name || user?.email || "U").slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuLabel className="truncate">
                {user?.display_name || user?.email}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => router.push("/app/profile")}>
                <User className="mr-2 h-4 w-4" /> Profile
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => router.push("/app/settings")}>
                Settings
              </DropdownMenuItem>
              {isAdmin ? (
                <DropdownMenuItem onClick={() => router.push("/admin")}>
                  <Shield className="mr-2 h-4 w-4" /> Admin portal
                </DropdownMenuItem>
              ) : null}
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => void logout()}>
                <LogOut className="mr-2 h-4 w-4" /> Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      }
    />
  );

  return (
    <>
      {mobileOpen ? (
        <div className="fixed inset-0 z-overlay lg:hidden">
          <button
            type="button"
            className="absolute inset-0 bg-black/50"
            aria-label="Close menu"
            onClick={() => setMobileOpen(false)}
          />
          <div className="absolute inset-y-0 left-0 w-72 border-r bg-sidebar shadow-cds-lg">
            {sidebar}
          </div>
        </div>
      ) : null}
      <DashboardShell sidebar={sidebar} topbar={topbar}>
        <div className="mb-4">
          <Breadcrumbs />
        </div>
        {children}
      </DashboardShell>
    </>
  );
}
