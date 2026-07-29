import type { LucideIcon } from "lucide-react";
import {
  LayoutDashboard,
  LineChart,
  Users,
  PiggyBank,
  GraduationCap,
  Radio,
  Wallet,
  Gift,
  Bell,
  Settings,
} from "lucide-react";

export interface AppNavItem {
  label: string;
  href: string;
  icon: LucideIcon;
}

/** Primary user portal navigation (authenticated shell). */
export const userNavItems: AppNavItem[] = [
  { label: "Dashboard", href: "/app/dashboard", icon: LayoutDashboard },
  { label: "Wallets", href: "/app/wallet", icon: Wallet },
  { label: "Earn", href: "/app/earn", icon: PiggyBank },
  { label: "Markets", href: "/app/markets", icon: LineChart },
  { label: "P2P", href: "/app/p2p", icon: Users },
  { label: "Signals", href: "/app/signals", icon: Radio },
  { label: "Referrals", href: "/app/referrals", icon: Gift },
  { label: "Academy", href: "/app/academy", icon: GraduationCap },
  { label: "Notifications", href: "/app/notifications", icon: Bell },
  { label: "Settings", href: "/app/settings", icon: Settings },
];
