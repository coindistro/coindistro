"use client";

import { useState, useEffect, useId } from "react";
import { motion, AnimatePresence } from "framer-motion";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu, X, Sun, Moon, LogOut, LayoutDashboard, Settings, User } from "lucide-react";
import {
  Avatar,
  AvatarFallback,
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  useTheme,
  cn,
} from "@coindistro/cds";
import { useAuth } from "@/features/authentication/auth-provider";
import { MobileNavDrawer } from "@/components/MobileNavDrawer";

const navLinks = [
  { name: "Products", href: "/#ecosystem" },
  { name: "Markets", href: "/#market" },
  { name: "Signals", href: "/#signals" },
  { name: "Academy", href: "/academy" },
  { name: "Security", href: "/#security" },
  { name: "Roadmap", href: "/#roadmap" },
];

function userInitials(displayName?: string | null, email?: string | null): string {
  const source = (displayName || email || "U").trim();
  return source.slice(0, 2).toUpperCase();
}

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const { resolvedTheme, setTheme } = useTheme();
  const { user, isAuthenticated, loading, logout } = useAuth();
  const pathname = usePathname();
  const drawerId = useId();
  const drawerDomId = `mobile-nav-${drawerId.replace(/:/g, "")}`;

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 50);
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  // Close drawer on route change (safety net for Link navigations).
  useEffect(() => {
    setMenuOpen(false);
  }, [pathname]);

  const isDark = resolvedTheme === "dark";
  const toggleTheme = () => setTheme(isDark ? "light" : "dark");

  const showGuestActions = !loading && !isAuthenticated;
  const showUserMenu = !loading && isAuthenticated;

  return (
    <motion.nav
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.6, ease: "easeOut" }}
      className={cn(
        "fixed top-0 left-0 right-0 z-50 transition-all duration-300",
        scrolled ? "glass py-3 shadow-lg" : "bg-transparent py-5",
      )}
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between gap-3">
          {/* Logo */}
          <Link href="/" className="flex min-w-0 items-center gap-2">
            <div className="relative h-8 w-8 shrink-0">
              <Image
                src="/coindistro-logo.png"
                alt="Coindistro Logo"
                fill
                className="object-contain"
                priority
              />
            </div>
            <span className="truncate text-xl font-bold gradient-text">Coindistro</span>
          </Link>

          {/* Desktop Nav */}
          <div className="hidden items-center gap-1 md:flex">
            {navLinks.map((link) => (
              <a
                key={link.name}
                href={link.href}
                className="rounded-lg px-4 py-2 text-sm text-[var(--text-muted)] transition-colors duration-200 hover:bg-white/5 hover:text-[var(--text-primary)]"
              >
                {link.name}
              </a>
            ))}
          </div>

          {/* Desktop: Theme + auth CTAs / avatar */}
          <div className="hidden items-center gap-3 md:flex">
            <button
              type="button"
              onClick={toggleTheme}
              className="group relative rounded-lg p-2.5 glass transition-all duration-300 hover:bg-[var(--card-bg)]/50"
              aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
            >
              <AnimatePresence mode="wait" initial={false}>
                {isDark ? (
                  <motion.div
                    key="sun"
                    initial={{ scale: 0, rotate: -180 }}
                    animate={{ scale: 1, rotate: 0 }}
                    exit={{ scale: 0, rotate: 180 }}
                    transition={{ duration: 0.3 }}
                  >
                    <Sun className="h-5 w-5 text-yellow-400" />
                  </motion.div>
                ) : (
                  <motion.div
                    key="moon"
                    initial={{ scale: 0, rotate: 180 }}
                    animate={{ scale: 1, rotate: 0 }}
                    exit={{ scale: 0, rotate: -180 }}
                    transition={{ duration: 0.3 }}
                  >
                    <Moon className="h-5 w-5 text-[#7C3AED]" />
                  </motion.div>
                )}
              </AnimatePresence>
            </button>

            {showGuestActions && (
              <>
                <Link
                  href="/login"
                  className="px-4 py-2 text-sm font-medium text-[var(--text-muted)] hover:text-[var(--text-primary)]"
                >
                  Login
                </Link>
                <Link
                  href="/register"
                  className="glow-purple rounded-lg bg-gradient-to-r from-[#7C3AED] to-[#06B6D4] px-5 py-2.5 text-sm font-medium text-[var(--text-primary)] transition-all duration-200 hover:opacity-90"
                >
                  Sign Up
                </Link>
              </>
            )}

            {showUserMenu && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    type="button"
                    aria-label="Account menu"
                    className="inline-flex items-center gap-2 rounded-full outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  >
                    <Avatar className="h-9 w-9 border border-border">
                      <AvatarFallback className="bg-primary/15 text-xs font-semibold text-primary">
                        {userInitials(user?.display_name, user?.email)}
                      </AvatarFallback>
                    </Avatar>
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" className="w-56">
                  <DropdownMenuLabel className="font-normal">
                    <div className="flex flex-col space-y-1">
                      <p className="text-sm font-medium leading-none">
                        {user?.display_name || "Account"}
                      </p>
                      <p className="truncate text-xs leading-none text-muted-foreground">
                        {user?.email}
                      </p>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem asChild>
                    <Link href="/app/dashboard" className="cursor-pointer">
                      <LayoutDashboard className="mr-2 h-4 w-4" />
                      Dashboard
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link href="/app/profile" className="cursor-pointer">
                      <User className="mr-2 h-4 w-4" />
                      Profile
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem asChild>
                    <Link href="/app/settings" className="cursor-pointer">
                      <Settings className="mr-2 h-4 w-4" />
                      Settings
                    </Link>
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="cursor-pointer text-destructive focus:text-destructive"
                    onSelect={() => {
                      void logout();
                    }}
                  >
                    <LogOut className="mr-2 h-4 w-4" />
                    Logout
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>

          {/* Mobile: Theme + hamburger */}
          <div className="flex items-center gap-1 sm:gap-2 md:hidden">
            <button
              type="button"
              onClick={toggleTheme}
              className="rounded-lg p-2 glass"
              aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
            >
              {isDark ? (
                <Sun className="h-5 w-5 text-yellow-400" />
              ) : (
                <Moon className="h-5 w-5 text-[#7C3AED]" />
              )}
            </button>

            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="text-[var(--text-primary)]"
              onClick={() => setMenuOpen((open) => !open)}
              aria-label={menuOpen ? "Close navigation menu" : "Open navigation menu"}
              aria-expanded={menuOpen}
              aria-controls={drawerDomId}
            >
              {menuOpen ? <X className="h-6 w-6" aria-hidden /> : <Menu className="h-6 w-6" aria-hidden />}
            </Button>
          </div>
        </div>
      </div>

      <MobileNavDrawer
        open={menuOpen}
        onOpenChange={setMenuOpen}
        navLinks={navLinks}
        isAuthenticated={isAuthenticated}
        authLoading={loading}
        onLogout={logout}
        drawerId={drawerDomId}
      />
    </motion.nav>
  );
}
