"use client";

import { Alert, AlertDescription, Button, Typography } from "@coindistro/cds";
import { AlertCircle } from "lucide-react";
import Link from "next/link";

interface AuthErrorScreenProps {
  title?: string;
  description: string;
  actionLabel?: string;
  actionHref?: string;
}

export function AuthErrorScreen({
  title = "Something went wrong",
  description,
  actionLabel = "Try again",
  actionHref = "/login",
}: AuthErrorScreenProps) {
  return (
    <div className="flex flex-col items-center justify-center py-8 text-center space-y-4">
      <div className="flex h-16 w-16 items-center justify-center rounded-full bg-destructive/20">
        <AlertCircle className="h-8 w-8 text-destructive" />
      </div>
      <Typography variant="h3">{title}</Typography>
      <Alert variant="danger" className="max-w-sm text-left">
        <AlertDescription>{description}</AlertDescription>
      </Alert>
      <Button asChild variant="outline">
        <Link href={actionHref}>{actionLabel}</Link>
      </Button>
    </div>
  );
}