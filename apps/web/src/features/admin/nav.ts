import type { LucideIcon } from "lucide-react";
import {
  LayoutDashboard,
  Users,
  Sparkles,
  Gift,
  Mail,
  ArrowLeftRight,
  PiggyBank,
  GraduationCap,
  Radio,
  Wallet,
  Store,
  LineChart,
  ScrollText,
  Flag,
  Cpu,
  CalendarClock,
  Activity,
  HeartPulse,
  Settings,
} from "lucide-react";

export interface AdminNavItem {
  label: string;
  href: string;
  icon: LucideIcon;
}

export const adminNavItems: AdminNavItem[] = [
  { label: "Overview", href: "/admin", icon: LayoutDashboard },
  { label: "Users", href: "/admin/users", icon: Users },
  { label: "Genesis Members", href: "/admin/genesis", icon: Sparkles },
  { label: "Referrals", href: "/admin/referrals", icon: Gift },
  { label: "Invitations", href: "/admin/invitations", icon: Mail },
  { label: "P2P", href: "/admin/p2p", icon: ArrowLeftRight },
  { label: "Earn", href: "/admin/earn", icon: PiggyBank },
  { label: "Academy", href: "/admin/academy", icon: GraduationCap },
  { label: "Signals", href: "/admin/signals", icon: Radio },
  { label: "Wallets", href: "/admin/wallets", icon: Wallet },
  { label: "Merchants", href: "/admin/merchants", icon: Store },
  { label: "Markets", href: "/admin/markets", icon: LineChart },
  { label: "Audit Logs", href: "/admin/audit", icon: ScrollText },
  { label: "Feature Flags", href: "/admin/feature-flags", icon: Flag },
  { label: "Workers", href: "/admin/workers", icon: Cpu },
  { label: "Scheduler", href: "/admin/scheduler", icon: CalendarClock },
  { label: "Metrics", href: "/admin/metrics", icon: Activity },
  { label: "System Health", href: "/admin/health", icon: HeartPulse },
  { label: "Settings", href: "/admin/settings", icon: Settings },
];
