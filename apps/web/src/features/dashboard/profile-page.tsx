"use client";

import * as React from "react";
import { useForm } from "react-hook-form";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Alert,
  AlertDescription,
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
  Field,
  FounderBadge,
  GenesisBadge,
  Input,
  PageHeader,
  Skeleton,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@coindistro/cds";
import { useAuth } from "@/features/authentication/auth-provider";
import * as identityApi from "@/features/identity/api";
import { useToast } from "@/features/shared/providers/toast-provider";
import { ApiError } from "@/lib/api/types";
import {
  displayName,
  formatDate,
  formatRelative,
  humanizeAction,
  initials,
  profileCompletion,
} from "@/lib/utils/format";

type ProfileForm = {
  display_name: string;
  username: string;
  country: string;
  timezone: string;
};

export function ProfilePage() {
  const { user, refreshUser } = useAuth();
  const { toast } = useToast();
  const qc = useQueryClient();

  const profileQ = useQuery({
    queryKey: ["users", "me"],
    queryFn: identityApi.getProfile,
  });
  const sessionsQ = useQuery({
    queryKey: ["sessions"],
    queryFn: identityApi.getSessions,
  });
  const devicesQ = useQuery({
    queryKey: ["devices"],
    queryFn: identityApi.getDevices,
  });
  const activityQ = useQuery({
    queryKey: ["activity"],
    queryFn: identityApi.getActivityLog,
  });
  const referralQ = useQuery({
    queryKey: ["referrals", "dashboard"],
    queryFn: identityApi.getReferralDashboard,
  });

  const me = profileQ.data ?? user;
  const completion = profileCompletion(me);

  const form = useForm<ProfileForm>({
    values: {
      display_name: me?.display_name || "",
      username: me?.username || "",
      country: me?.country || "",
      timezone: me?.timezone || "UTC",
    },
  });

  const updateMut = useMutation({
    mutationFn: identityApi.updateProfile,
    onSuccess: async () => {
      toast({ message: "Profile updated", variant: "success" });
      await qc.invalidateQueries({ queryKey: ["users", "me"] });
      await refreshUser();
    },
    onError: (e) => {
      toast({
        message: e instanceof ApiError ? e.message : "Update failed",
        variant: "danger",
      });
    },
  });

  const terminateMut = useMutation({
    mutationFn: identityApi.terminateSession,
    onSuccess: async () => {
      toast({ message: "Session terminated", variant: "success" });
      await qc.invalidateQueries({ queryKey: ["sessions"] });
    },
  });

  const onSubmit = form.handleSubmit((values) => {
    updateMut.mutate({
      display_name: values.display_name || undefined,
      username: values.username || undefined,
      country: values.country || undefined,
      timezone: values.timezone || undefined,
    });
  });

  if (profileQ.isLoading && !me) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-cds-fade-in">
      <PageHeader
        title="Profile"
        description="Identity, referral information, sessions, devices, and security."
        actions={
          <div className="flex flex-wrap gap-2">
            {me?.is_genesis ? <GenesisBadge number={me.genesis_number ?? undefined} /> : null}
            {me?.is_founder ? <FounderBadge /> : null}
            <Badge variant={me?.is_verified ? "success" : "warning"}>
              {me?.is_verified ? "Verified" : "Unverified"}
            </Badge>
          </div>
        }
      />

      <Card>
        <CardContent className="flex flex-col gap-4 p-6 sm:flex-row sm:items-center">
          <Avatar className="h-16 w-16">
            {me?.avatar_url ? (
              <AvatarImage src={me.avatar_url} alt={displayName(me)} />
            ) : null}
            <AvatarFallback className="text-lg">{initials(me)}</AvatarFallback>
          </Avatar>
          <div className="flex-1">
            <p className="text-lg font-semibold">{displayName(me)}</p>
            <p className="text-sm text-muted-foreground">{me?.email}</p>
            <p className="mt-1 text-xs text-muted-foreground">
              Member since {formatDate(me?.created_at)} · Profile {completion.percent}% complete
            </p>
          </div>
        </CardContent>
      </Card>

      <Tabs defaultValue="profile">
        <TabsList className="flex h-auto flex-wrap">
          <TabsTrigger value="profile">Profile</TabsTrigger>
          <TabsTrigger value="referral">Referral</TabsTrigger>
          <TabsTrigger value="sessions">Sessions</TabsTrigger>
          <TabsTrigger value="devices">Devices</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
          <TabsTrigger value="activity">Activity</TabsTrigger>
        </TabsList>

        <TabsContent value="profile" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Edit profile</CardTitle>
              <CardDescription>Updates sync via GET/PUT /users/me</CardDescription>
            </CardHeader>
            <CardContent>
              <form className="grid max-w-xl gap-4" onSubmit={onSubmit}>
                <Field label="Display name" htmlFor="display_name">
                  <Input id="display_name" {...form.register("display_name")} />
                </Field>
                <Field label="Username" htmlFor="username">
                  <Input id="username" {...form.register("username")} />
                </Field>
                <Field
                  label="Country"
                  htmlFor="country"
                  description="ISO 3166-1 alpha-3 (e.g. NGA)"
                >
                  <Input id="country" maxLength={3} {...form.register("country")} />
                </Field>
                <Field label="Timezone" htmlFor="timezone">
                  <Input id="timezone" {...form.register("timezone")} />
                </Field>
                <Button type="submit" loading={updateMut.isPending}>
                  Save changes
                </Button>
              </form>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="referral" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Referral information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              {referralQ.isLoading ? (
                <Skeleton className="h-24 w-full" />
              ) : referralQ.data ? (
                <>
                  <div className="grid gap-3 sm:grid-cols-2">
                    <div>
                      <p className="text-xs text-muted-foreground">Your code</p>
                      <p className="font-mono text-lg font-semibold">
                        {referralQ.data.referral_code}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Invitation credits</p>
                      <p className="text-lg font-semibold">
                        {referralQ.data.invitation_credits}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Successful invites</p>
                      <p className="font-semibold">{referralQ.data.successful_invites}</p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Conversion rate</p>
                      <p className="font-semibold">{referralQ.data.conversion_rate}%</p>
                    </div>
                  </div>
                  {referralQ.data.referral_link ? (
                    <p className="break-all text-muted-foreground">
                      Link: {referralQ.data.referral_link}
                    </p>
                  ) : null}
                </>
              ) : (
                <Alert variant="danger">
                  <AlertDescription>Could not load referral data.</AlertDescription>
                </Alert>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="sessions" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Active sessions</CardTitle>
              <CardDescription>Devices currently signed in</CardDescription>
            </CardHeader>
            <CardContent>
              {sessionsQ.isLoading ? (
                <Skeleton className="h-32 w-full" />
              ) : !sessionsQ.data?.length ? (
                <p className="text-sm text-muted-foreground">No sessions found.</p>
              ) : (
                <ul className="divide-y divide-border/60">
                  {sessionsQ.data.map((s) => (
                    <li
                      key={s.id}
                      className="flex flex-col gap-2 py-3 sm:flex-row sm:items-center sm:justify-between"
                    >
                      <div className="text-sm">
                        <p className="font-medium">
                          {s.browser || s.device_name || "Session"}
                          {s.is_current ? (
                            <Badge variant="success" className="ml-2 text-[10px]">
                              Current
                            </Badge>
                          ) : null}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {s.operating_system || "Unknown OS"}
                          {s.ip_address ? ` · ${s.ip_address}` : ""}
                          {" · "}
                          {formatRelative(s.last_activity_at)}
                        </p>
                      </div>
                      {!s.is_current ? (
                        <Button
                          variant="outline"
                          size="sm"
                          loading={terminateMut.isPending}
                          onClick={() => terminateMut.mutate(s.id)}
                        >
                          Terminate
                        </Button>
                      ) : null}
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="devices" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Devices</CardTitle>
            </CardHeader>
            <CardContent>
              {devicesQ.isLoading ? (
                <Skeleton className="h-32 w-full" />
              ) : !devicesQ.data?.length ? (
                <p className="text-sm text-muted-foreground">No trusted devices recorded.</p>
              ) : (
                <ul className="divide-y divide-border/60">
                  {devicesQ.data.map((d) => (
                    <li key={d.id} className="py-3 text-sm">
                      <p className="font-medium">
                        {d.name || d.browser || "Device"}
                        {d.is_current ? (
                          <Badge variant="info" className="ml-2 text-[10px]">
                            Current
                          </Badge>
                        ) : null}
                        {d.is_trusted ? (
                          <Badge variant="success" className="ml-2 text-[10px]">
                            Trusted
                          </Badge>
                        ) : null}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {d.operating_system || "—"} · Last seen {formatRelative(d.last_seen_at)}
                      </p>
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="security" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Security</CardTitle>
              <CardDescription>Account protection status</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex justify-between border-b border-border/50 py-2">
                <span className="text-muted-foreground">Email verification</span>
                <span>{me?.is_verified ? "Verified" : "Pending"}</span>
              </div>
              <div className="flex justify-between border-b border-border/50 py-2">
                <span className="text-muted-foreground">Account status</span>
                <span className="capitalize">{me?.status ?? "—"}</span>
              </div>
              <div className="flex justify-between border-b border-border/50 py-2">
                <span className="text-muted-foreground">Last login</span>
                <span>{formatDate(me?.last_login_at)}</span>
              </div>
              <div className="flex justify-between py-2">
                <span className="text-muted-foreground">Roles</span>
                <span>{(me?.roles || ["user"]).join(", ")}</span>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="activity" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Activity</CardTitle>
            </CardHeader>
            <CardContent>
              {activityQ.isLoading ? (
                <Skeleton className="h-40 w-full" />
              ) : !activityQ.data?.length ? (
                <p className="text-sm text-muted-foreground">No activity yet.</p>
              ) : (
                <ul className="divide-y divide-border/60">
                  {activityQ.data.map((a) => (
                    <li key={a.id} className="flex justify-between gap-4 py-2 text-sm">
                      <span className="font-medium">{humanizeAction(a.action)}</span>
                      <span className="shrink-0 text-xs text-muted-foreground">
                        {formatRelative(a.created_at)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
