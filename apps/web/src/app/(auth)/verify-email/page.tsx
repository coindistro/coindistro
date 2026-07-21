"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Alert, AlertDescription, Button, Spinner, Typography } from "@coindistro/cds";
import { verifyEmail } from "@/features/authentication/api";
import { ApiError } from "@/lib/api/types";

export default function VerifyEmailPage() {
  const params = useSearchParams();
  const token = params.get("token") || "";
  const [status, setStatus] = useState<"loading" | "ok" | "error">("loading");
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (!token) {
      setStatus("error");
      setMessage("Missing verification token.");
      return;
    }
    verifyEmail(token)
      .then(() => setStatus("ok"))
      .catch((e) => {
        setStatus("error");
        setMessage(e instanceof ApiError ? e.message : "Verification failed");
      });
  }, [token]);

  if (status === "loading") {
    return (
      <div className="flex flex-col items-center gap-3 py-8">
        <Spinner label="Verifying email" />
        <Typography variant="body">Verifying your email…</Typography>
      </div>
    );
  }

  if (status === "ok") {
    return (
      <div className="space-y-4 text-center">
        <Typography variant="h3">Email verified</Typography>
        <p className="text-sm text-muted-foreground">Your account is ready.</p>
        <Button asChild>
          <Link href="/app/dashboard">Go to dashboard</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <Typography variant="h3">Verification failed</Typography>
      <Alert variant="danger">
        <AlertDescription>{message}</AlertDescription>
      </Alert>
      <Button asChild variant="outline">
        <Link href="/login">Sign in</Link>
      </Button>
    </div>
  );
}
