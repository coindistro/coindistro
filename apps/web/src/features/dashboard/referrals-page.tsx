"use client";

import * as React from "react";
import { useForm } from "react-hook-form";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Field,
  Input,
  PageHeader,
  Skeleton,
  StatCard,
  Textarea,
} from "@coindistro/cds";
import { Copy, Gift, Mail, Users } from "lucide-react";
import * as identityApi from "@/features/identity/api";
import { useToast } from "@/features/shared/providers/toast-provider";
import { ApiError } from "@/lib/api/types";
import { formatRelative } from "@/lib/utils/format";

export function ReferralsPage() {
  const { toast } = useToast();
  const qc = useQueryClient();
  const dashboardQ = useQuery({
    queryKey: ["referrals", "dashboard"],
    queryFn: identityApi.getReferralDashboard,
  });
  const invitesQ = useQuery({
    queryKey: ["invitations"],
    queryFn: identityApi.getInvitations,
  });

  const form = useForm<{ email: string; message: string }>({
    defaultValues: { email: "", message: "" },
  });

  const inviteMut = useMutation({
    mutationFn: (v: { email: string; message?: string }) =>
      identityApi.sendInvitation(v.email, v.message),
    onSuccess: async () => {
      toast({ message: "Invitation sent", variant: "success" });
      form.reset();
      await qc.invalidateQueries({ queryKey: ["invitations"] });
      await qc.invalidateQueries({ queryKey: ["referrals", "dashboard"] });
    },
    onError: (e) => {
      toast({
        message: e instanceof ApiError ? e.message : "Failed to send invitation",
        variant: "danger",
      });
    },
  });

  const d = dashboardQ.data;

  const copy = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast({ message: `${label} copied`, variant: "success" });
    } catch {
      toast({ message: "Copy failed", variant: "danger" });
    }
  };

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title="Referrals"
        description="Invite friends, track credits, and grow your network."
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {dashboardQ.isLoading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <Skeleton className="h-12 w-full" />
              </CardContent>
            </Card>
          ))
        ) : (
          <>
            <StatCard
              title="Credits"
              value={d?.invitation_credits ?? 0}
              icon={<Mail className="h-4 w-4" />}
            />
            <StatCard
              title="Total invites"
              value={d?.total_invites ?? 0}
              icon={<Users className="h-4 w-4" />}
            />
            <StatCard
              title="Successful"
              value={d?.successful_invites ?? 0}
              icon={<Gift className="h-4 w-4" />}
            />
            <StatCard
              title="Conversion"
              value={`${d?.conversion_rate ?? 0}%`}
              description={`Rewards: ${d?.rewards_earned ?? 0}`}
            />
          </>
        )}
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Your referral code</CardTitle>
            <CardDescription>Share this code so friends can register</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded-md border bg-muted/40 px-3 py-2 font-mono text-lg">
                {d?.referral_code || "—"}
              </code>
              <Button
                variant="outline"
                size="icon"
                aria-label="Copy code"
                onClick={() => d?.referral_code && void copy(d.referral_code, "Code")}
              >
                <Copy className="h-4 w-4" />
              </Button>
            </div>
            {d?.referral_link ? (
              <div className="flex items-center gap-2">
                <p className="flex-1 truncate text-sm text-muted-foreground">{d.referral_link}</p>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => void copy(d.referral_link, "Link")}
                >
                  Copy link
                </Button>
              </div>
            ) : null}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Send invitation</CardTitle>
            <CardDescription>Uses one invitation credit</CardDescription>
          </CardHeader>
          <CardContent>
            <form
              className="space-y-3"
              onSubmit={form.handleSubmit((v) =>
                inviteMut.mutate({
                  email: v.email,
                  message: v.message || undefined,
                }),
              )}
            >
              <Field label="Email" htmlFor="invite-email" required>
                <Input
                  id="invite-email"
                  type="email"
                  {...form.register("email", { required: true })}
                />
              </Field>
              <Field label="Message" htmlFor="invite-message">
                <Textarea id="invite-message" rows={3} {...form.register("message")} />
              </Field>
              <Button type="submit" loading={inviteMut.isPending}>
                Send invite
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Invitation history</CardTitle>
        </CardHeader>
        <CardContent>
          {invitesQ.isLoading ? (
            <Skeleton className="h-32 w-full" />
          ) : !invitesQ.data?.length ? (
            <p className="text-sm text-muted-foreground">No invitations sent yet.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="text-xs text-muted-foreground">
                  <tr className="border-b">
                    <th className="pb-2 font-medium">Email</th>
                    <th className="pb-2 font-medium">Status</th>
                    <th className="pb-2 font-medium">Code</th>
                    <th className="pb-2 font-medium">Sent</th>
                  </tr>
                </thead>
                <tbody>
                  {invitesQ.data.map((inv) => (
                    <tr key={inv.id} className="border-b border-border/50 last:border-0">
                      <td className="py-2">{inv.invitee_email}</td>
                      <td className="py-2 capitalize">{inv.status}</td>
                      <td className="py-2 font-mono text-xs">{inv.code}</td>
                      <td className="py-2 text-muted-foreground">
                        {formatRelative(inv.created_at)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
