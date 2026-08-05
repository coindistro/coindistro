"use client";

import * as React from "react";
import Link from "next/link";
import { CheckCircle2, Shield, UserRound } from "lucide-react";
import { Badge, Progress } from "@coindistro/cds";
import { SectionShell, GlassCard } from "./section-shell";

export interface AccountStatusData {
  verified?: boolean;
  twoFactorEnabled?: boolean;
  kycLevel?: string;
  membership?: string;
  referralTier?: string;
  securityScore?: number;
  profileCompletion?: number;
  missingProfile?: string[];
}

export function AccountStatusSection({ data }: { data: AccountStatusData }) {
  const security = Math.min(100, Math.max(0, data.securityScore ?? 0));
  const profile = Math.min(100, Math.max(0, data.profileCompletion ?? 0));

  return (
    <SectionShell
      id="account-status"
      title="Account Status"
      description="Verification, security, and membership"
      actionHref="/app/profile"
      actionLabel="Profile"
    >
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
        <GlassCard>
          <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
            <CheckCircle2 className="h-4 w-4 text-primary" aria-hidden />
            Verification
          </div>
          <div className="mt-3 space-y-2 text-sm">
            <Row
              label="Email"
              value={
                <Badge variant={data.verified ? "success" : "warning"}>
                  {data.verified ? "Verified" : "Unverified"}
                </Badge>
              }
            />
            <Row label="KYC Level" value={data.kycLevel || "Level 0"} />
            <Row
              label="2FA"
              value={
                <Badge variant={data.twoFactorEnabled ? "success" : "outline"}>
                  {data.twoFactorEnabled ? "Enabled" : "Not enabled"}
                </Badge>
              }
            />
          </div>
        </GlassCard>

        <GlassCard>
          <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
            <UserRound className="h-4 w-4 text-primary" aria-hidden />
            Membership
          </div>
          <div className="mt-3 space-y-2 text-sm">
            <Row label="Membership" value={data.membership || "Standard"} />
            <Row label="Referral Tier" value={data.referralTier || "Starter"} />
            <div>
              <div className="mb-1 flex justify-between text-xs text-muted-foreground">
                <span>Profile completion</span>
                <span>{profile}%</span>
              </div>
              <Progress value={profile} />
              {data.missingProfile?.length ? (
                <p className="mt-1.5 text-[11px] text-muted-foreground">
                  Missing: {data.missingProfile.slice(0, 2).join(", ")}
                </p>
              ) : null}
            </div>
          </div>
        </GlassCard>

        <GlassCard className="sm:col-span-2 xl:col-span-1">
          <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
            <Shield className="h-4 w-4 text-primary" aria-hidden />
            Security Score
          </div>
          <p className="mt-3 text-3xl font-bold tabular-nums text-foreground">{security}</p>
          <Progress value={security} className="mt-2" />
          <p className="mt-2 text-xs text-muted-foreground">
            Based on email verification, session activity, and profile strength.
          </p>
          <Link
            href="/app/settings"
            className="mt-3 inline-flex text-xs font-semibold text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            Improve security →
          </Link>
        </GlassCard>
      </div>
    </SectionShell>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-2">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium text-foreground">{value}</span>
    </div>
  );
}
