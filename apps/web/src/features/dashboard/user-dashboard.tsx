"use client";

import * as React from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  FounderBadge,
  GenesisBadge,
  PageHeader,
  Progress,
  Skeleton,
  StatCard,
  StatusDot,
  Typography,
} from "@coindistro/cds";
import {
  ArrowRight,
  Bell,
  Copy,
  Gift,
  PiggyBank,
  Shield,
  Sparkles,
  User,
  Wallet,
} from "lucide-react";
import { useAuth } from "@/features/authentication/auth-provider";
import * as identityApi from "@/features/identity/api";
import { getEarnPortfolio } from "@/features/earn/api";
import { useToast } from "@/features/shared/providers/toast-provider";
import {
  displayName,
  formatRelative,
  humanizeAction,
  initials,
  profileCompletion,
} from "@/lib/utils/format";

function PlaceholderMark({ label = "Placeholder" }: { label?: string }) {
  return (
    <Badge variant="outline" className="text-[10px] font-normal text-muted-foreground">
      {label}
    </Badge>
  );
}

export function UserDashboard() {
  const { user } = useAuth();
  const { toast } = useToast();

  const profileQ = useQuery({
    queryKey: ["users", "me"],
    queryFn: identityApi.getProfile,
    initialData: user ?? undefined,
  });
  const referralQ = useQuery({
    queryKey: ["referrals", "dashboard"],
    queryFn: identityApi.getReferralDashboard,
  });
  const activityQ = useQuery({
    queryKey: ["activity"],
    queryFn: identityApi.getActivityLog,
  });
  const sessionsQ = useQuery({
    queryKey: ["sessions"],
    queryFn: identityApi.getSessions,
  });
  const earnQ = useQuery({
    queryKey: ["earn", "portfolio"],
    queryFn: getEarnPortfolio,
  });

  const me = profileQ.data ?? user;
  const name = displayName(me);
  const completion = profileCompletion(me);
  const referrals = referralQ.data;
  const activity = activityQ.data ?? [];
  const sessions = sessionsQ.data ?? [];
  const earn = earnQ.data;
  const currentSession = sessions.find((s) => s.is_current) ?? sessions[0];

  const copyReferral = async () => {
    const code = referrals?.referral_code || me?.referral_code;
    if (!code) return;
    try {
      await navigator.clipboard.writeText(code);
      toast({ message: "Referral code copied", variant: "success" });
    } catch {
      toast({ message: "Could not copy code", variant: "danger" });
    }
  };

  const loading =
    profileQ.isLoading || referralQ.isLoading || activityQ.isLoading;

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title={`Welcome back, ${name}`}
        description="Your live Coindistro overview — identity, earn, referrals, and security."
        actions={
          <div className="flex flex-wrap items-center gap-2">
            {me?.is_genesis ? <GenesisBadge number={me.genesis_number ?? undefined} /> : null}
            {me?.is_founder ? <FounderBadge /> : null}
            {me?.is_verified ? (
              <Badge variant="success">Verified</Badge>
            ) : (
              <Badge variant="warning">Email unverified</Badge>
            )}
          </div>
        }
      />

      {/* Identity strip */}
      <Card>
        <CardContent className="flex flex-col gap-4 p-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-4">
            <Avatar className="h-14 w-14">
              {me?.avatar_url ? <AvatarImage src={me.avatar_url} alt={name} /> : null}
              <AvatarFallback>{initials(me)}</AvatarFallback>
            </Avatar>
            <div>
              <Typography variant="h4" className="text-lg">
                {name}
              </Typography>
              <p className="text-sm text-muted-foreground">{me?.email}</p>
              <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                <span className="inline-flex items-center gap-1">
                  <StatusDot status={me?.status === "active" ? "online" : "busy"} />
                  {me?.status ?? "unknown"}
                </span>
                {me?.roles?.map((r) => (
                  <Badge key={r} variant="secondary" className="text-[10px] capitalize">
                    {r.replace("_", " ")}
                  </Badge>
                ))}
              </div>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" asChild>
              <Link href="/app/profile">
                <User className="mr-2 h-4 w-4" /> Profile
              </Link>
            </Button>
            <Button variant="outline" size="sm" asChild>
              <Link href="/app/referrals">
                <Gift className="mr-2 h-4 w-4" /> Referrals
              </Link>
            </Button>
            <Button size="sm" asChild>
              <Link href="/app/earn">
                <PiggyBank className="mr-2 h-4 w-4" /> Earn
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Stats row */}
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {loading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <Card key={i}>
              <CardContent className="space-y-3 p-6">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-8 w-16" />
              </CardContent>
            </Card>
          ))
        ) : (
          <>
            <StatCard
              title="Genesis status"
              value={me?.is_genesis ? `Member #${me.genesis_number ?? "—"}` : "Not enrolled"}
              description={me?.is_genesis ? "Active genesis member" : "Open slots may apply"}
              icon={<Sparkles className="h-4 w-4" />}
            />
            <StatCard
              title="Referral code"
              value={
                <button
                  type="button"
                  onClick={() => void copyReferral()}
                  className="inline-flex items-center gap-2 hover:text-primary"
                  title="Copy code"
                >
                  <span className="font-mono text-xl">
                    {referrals?.referral_code || me?.referral_code || "—"}
                  </span>
                  <Copy className="h-4 w-4 text-muted-foreground" />
                </button>
              }
              description="Share to invite friends"
              icon={<Gift className="h-4 w-4" />}
            />
            <StatCard
              title="Invitation credits"
              value={referrals?.invitation_credits ?? "—"}
              description={
                referralQ.isError
                  ? "Unavailable"
                  : `${referrals?.pending_invites ?? 0} pending invites`
              }
              icon={<Bell className="h-4 w-4" />}
            />
            <StatCard
              title="Profile completion"
              value={`${completion.percent}%`}
              description={
                completion.missing.length
                  ? `Missing: ${completion.missing.slice(0, 2).join(", ")}`
                  : "Complete"
              }
              icon={<User className="h-4 w-4" />}
            />
          </>
        )}
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        {/* Portfolio / Earn */}
        <Card className="lg:col-span-1">
          <CardHeader className="flex flex-row items-start justify-between space-y-0">
            <div>
              <CardTitle className="text-base">Portfolio</CardTitle>
              <CardDescription>Earn balances and rewards</CardDescription>
            </div>
            {!earn ? <PlaceholderMark label="Wallet soon" /> : null}
          </CardHeader>
          <CardContent className="space-y-4">
            {earnQ.isLoading ? (
              <Skeleton className="h-28 w-full" />
            ) : earn ? (
              <>
                <div>
                  <p className="text-xs text-muted-foreground">Assets in Earn</p>
                  <p className="text-2xl font-bold tabular-nums">
                    {earn.total_assets_in_earn.toLocaleString(undefined, {
                      maximumFractionDigits: 4,
                    })}
                  </p>
                </div>
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <p className="text-xs text-muted-foreground">Today</p>
                    <p className="font-semibold tabular-nums">{earn.todays_rewards}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Lifetime</p>
                    <p className="font-semibold tabular-nums">{earn.lifetime_rewards}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Active products</p>
                    <p className="font-semibold">{earn.active_products}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Locked</p>
                    <p className="font-semibold tabular-nums">{earn.locked_balance}</p>
                  </div>
                </div>
              </>
            ) : (
              <div className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
                <Wallet className="mb-2 h-5 w-5" />
                Portfolio and wallet balances will appear when Earn / Wallet APIs are active.
                Spot balances are not available in this milestone.
              </div>
            )}
          </CardContent>
        </Card>

        {/* Earn summary + referrals */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="text-base">Earn summary</CardTitle>
            <CardDescription>Rewards at a glance</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {earn ? (
              <>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Estimated rewards</span>
                  <span className="font-medium tabular-nums">{earn.estimated_rewards}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Lifetime rewards</span>
                  <span className="font-medium tabular-nums">{earn.lifetime_rewards}</span>
                </div>
                <Button variant="outline" size="sm" className="w-full" asChild>
                  <Link href="/app/earn">
                    Open Earn <ArrowRight className="ml-2 h-4 w-4" />
                  </Link>
                </Button>
              </>
            ) : (
              <div className="space-y-2 text-sm text-muted-foreground">
                <p>Connect to Earn products to start earning yield and launchpool rewards.</p>
                <Button variant="outline" size="sm" asChild>
                  <Link href="/app/earn">Explore Earn</Link>
                </Button>
              </div>
            )}
            <div className="border-t pt-3">
              <p className="mb-2 text-sm font-medium">Referral statistics</p>
              {referralQ.isLoading ? (
                <Skeleton className="h-16 w-full" />
              ) : referrals ? (
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <p className="text-xs text-muted-foreground">Total invites</p>
                    <p className="font-semibold">{referrals.total_invites}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Successful</p>
                    <p className="font-semibold">{referrals.successful_invites}</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Conversion</p>
                    <p className="font-semibold">{referrals.conversion_rate}%</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Rewards</p>
                    <p className="font-semibold">{referrals.rewards_earned}</p>
                  </div>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">Referral data unavailable.</p>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Activity + security */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="text-base">Latest activity</CardTitle>
            <CardDescription>Security and account events</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {activityQ.isLoading ? (
              <Skeleton className="h-32 w-full" />
            ) : activity.length === 0 ? (
              <p className="text-sm text-muted-foreground">No recent activity yet.</p>
            ) : (
              <ul className="space-y-3">
                {activity.slice(0, 5).map((a) => (
                  <li key={a.id} className="flex items-start justify-between gap-2 text-sm">
                    <div>
                      <p className="font-medium">{humanizeAction(a.action)}</p>
                      <p className="text-xs text-muted-foreground">
                        {a.ip_address ? `${a.ip_address} · ` : ""}
                        {formatRelative(a.created_at)}
                      </p>
                    </div>
                  </li>
                ))}
              </ul>
            )}
            <Button variant="ghost" size="sm" className="w-full" asChild>
              <Link href="/app/notifications">View all notifications</Link>
            </Button>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">Recent login</CardTitle>
          </CardHeader>
          <CardContent className="space-y-1 text-sm">
            <p className="font-medium">
              {formatRelative(me?.last_login_at || currentSession?.login_at)}
            </p>
            <p className="text-muted-foreground">
              {currentSession?.browser || currentSession?.device_name || "Current device"}
              {currentSession?.ip_address ? ` · ${currentSession.ip_address}` : ""}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-base">Notifications</CardTitle>
            <PlaceholderMark label="Preview" />
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p>Successful login, security alerts, referrals, and system notices appear here.</p>
            <Button variant="outline" size="sm" asChild>
              <Link href="/app/notifications">Open inbox</Link>
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-base">Market snapshot</CardTitle>
            <PlaceholderMark />
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            Live market data ships with the Markets module. Navigation is ready under Markets.
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">Security status</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-2 text-sm">
              <Shield className="h-4 w-4 text-success" />
              <span>Session active</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <StatusDot status={me?.is_verified ? "online" : "pending"} />
              <span>{me?.is_verified ? "Email verified" : "Verify your email"}</span>
            </div>
            <div>
              <div className="mb-1 flex justify-between text-xs text-muted-foreground">
                <span>Profile</span>
                <span>{completion.percent}%</span>
              </div>
              <Progress value={completion.percent} />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick actions */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Quick actions</CardTitle>
          <CardDescription>Jump into the modules available today</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          {[
            { href: "/app/profile", label: "Edit profile" },
            { href: "/app/referrals", label: "Invite friends" },
            { href: "/app/settings", label: "Settings" },
            { href: "/app/earn", label: "Browse Earn" },
            { href: "/app/notifications", label: "Notifications" },
          ].map((a) => (
            <Button key={a.href} variant="secondary" size="sm" asChild>
              <Link href={a.href}>{a.label}</Link>
            </Button>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
