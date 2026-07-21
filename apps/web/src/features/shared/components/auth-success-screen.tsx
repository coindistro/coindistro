"use client";

import { Button, Typography } from "@coindistro/cds";
import { CheckCircle2 } from "lucide-react";
import Link from "next/link";

interface AuthSuccessScreenProps {
  title?: string;
  description?: string;
  actionLabel?: string;
  actionHref?: string;
}

export function AuthSuccessScreen({
  title = "Success",
  description = "Your request was completed successfully.",
  actionLabel = "Continue",
  actionHref = "/app/dashboard",
}: AuthSuccessScreenProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center space-y-4">
      <div className="flex h-16 w-16 items-center justify-center rounded-full bg-success/20">
        <CheckCircle2 className="h-8 w-8 text-success" />
      </div>
      <Typography variant="h3">{title}</Typography>
      <p className="max-w-sm text-sm text-muted-foreground">{description}</p>
      <Button asChild>
        <Link href={actionHref}>{actionLabel}</Link>
      </Button>
    </div>
  );
}