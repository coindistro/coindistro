"use client";

import Link from "next/link";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  PageHeader,
} from "@coindistro/cds";
import { useAuth } from "@/features/authentication/auth-provider";
import { displayName } from "@/lib/utils/format";

export default function MerchantHomePage() {
  const { user, logout } = useAuth();

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6 animate-cds-fade-in">
      <PageHeader
        title="Merchant portal"
        description="Accept crypto payments and manage merchant settlements."
        actions={<Badge variant="info">Merchant</Badge>}
      />

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Welcome, {displayName(user)}</CardTitle>
          <CardDescription>
            You are signed in as {user?.email}. Full merchant tooling ships in a later milestone.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          <Button asChild variant="secondary">
            <Link href="/dashboard">User dashboard</Link>
          </Button>
          <Button asChild variant="outline">
            <Link href="/app/settings">Settings</Link>
          </Button>
          <Button variant="ghost" onClick={() => void logout()}>
            Log out
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
