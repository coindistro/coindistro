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
  SearchInput,
  SidebarNav,
  Topbar,
  Badge,
  useTheme,
} from "@coindistro/cds";
import { LogOut, Moon, Shield, Sun } from "lucide-react";
import { useAuth } from "@/features/authentication/auth-provider";
import { adminNavItems } from "@/features/admin/nav";

export function AdminChrome({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout, isAdmin } = useAuth();
  const roleLabel =
    user?.roles?.find((r) => ["super_admin", "admin", "moderator"].includes(r)) ||
    (isAdmin ? "admin" : "staff");
  const { theme, setTheme } = useTheme();
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const [search, setSearch] = React.useState("");

  const items = adminNavItems.map((item) => ({
    label: item.label,
    href: item.href,
    icon: <item.icon className="h-4 w-4" />,
    active:
      item.href === "/admin"
        ? pathname === "/admin"
        : pathname === item.href || pathname.startsWith(item.href + "/"),
  }));

  const header = (
    <Link href="/admin" className="flex items-center gap-2 px-1">
      <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-destructive/90 text-sm font-bold text-white">
        <Shield className="h-4 w-4" />
      </span>
      <div>
        <div className="text-sm font-semibold leading-none">Admin</div>
        <div className="text-[10px] text-muted-foreground">Coindistro Control</div>
      </div>
    </Link>
  );

  const sidebar = (
    <SidebarNav
      header={header}
      items={items}
      onNavigate={(href) => {
        setMobileOpen(false);
        router.push(href);
      }}
    />
  );

  const topbar = (
    <Topbar
      title="Administration"
      onMenuClick={() => setMobileOpen((o) => !o)}
      search={
        <SearchInput
          placeholder="Search admin…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onClear={() => setSearch("")}
        />
      }
      actions={
        <div className="flex items-center gap-2">
          <Badge variant="danger" className="capitalize">
            {roleLabel.replace("_", " ")}
          </Badge>
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
              <Button variant="ghost" size="icon-sm" aria-label="Admin menu">
                <Avatar className="h-7 w-7">
                  <AvatarFallback className="text-xs">AD</AvatarFallback>
                </Avatar>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>{user?.email}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => router.push("/app/dashboard")}>
                User portal
              </DropdownMenuItem>
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
        {children}
      </DashboardShell>
    </>
  );
}
