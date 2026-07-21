import type { LucideIcon } from "lucide-react";
import {
  LayoutDashboard,
  LineChart,
  ArrowLeftRight,
  Users,
  PiggyBank,
  GraduationCap,
  Radio,
  Bot,
  Wallet,
  Store,
  CreditCard,
  Gift,
  Bell,
  User,
  Settings,
} from "lucide-react";

export interface AppNavItem {
  label: string;
  href: string;
  icon: LucideIcon;
}

export const userNavItems: AppNavItem[] = [
  { label: "Dashboard", href: "/app/dashboard", icon: LayoutDashboard },
  { label: "Markets", href: "/app/markets", icon: LineChart },
  { label: "Trade", href: "/app/trade", icon: ArrowLeftRight },
  { label: "P2P", href: "/app/p2p", icon: Users },
  { label: "Earn", href: "/app/earn", icon: PiggyBank },
  { label: "Academy", href: "/app/academy", icon: GraduationCap },
  { label: "Signals", href: "/app/signals", icon: Radio },
  { label: "AI Bots", href: "/app/ai-bots", icon: Bot },
  { label: "Wallet", href: "/app/wallet", icon: Wallet },
  { label: "Merchant", href: "/app/merchant", icon: Store },
  { label: "Pay", href: "/app/pay", icon: CreditCard },
  { label: "Referrals", href: "/app/referrals", icon: Gift },
  { label: "Notifications", href: "/app/notifications", icon: Bell },
  { label: "Profile", href: "/app/profile", icon: User },
  { label: "Settings", href: "/app/settings", icon: Settings },
];
