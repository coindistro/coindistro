"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import {
  Alert,
  AlertDescription,
  Button,
  Spinner,
  Typography,
} from "@coindistro/cds";
import { verifyEmail, resendVerification } from "@/features/authentication/api";
import { mapAuthError } from "@/features/authentication/auth-errors";
import { useAuth } from "@/features/authentication/auth-provider";
import { useToast } from "@/features/shared/providers/toast-provider";

export default function VerifyEmailPage() {
  const params = useSearchParams();
  const token = params.get("token") || "";
  const pending = params.get("pending") === "1";
  const email = params.get("email") || "";
  const { isAuthenticated } = useAuth();
  const { toast } = useToast();
  const [status, setStatus] = useState<"loading" | "ok" | "error" | "pending">(
    token ? "loading" : pending ? "pending" : "error",
  );
  const [message, setMessage] = useState(
    token
      ? ""
      : pending
        ? "Check your inbox for a verification link. You can resend it below if needed."
        : "Missing verification token.",
  );
  const [resending, setResending] = useState(false);

  useEffect(() => {
    if (!token) return;
    verifyEmail({ token })
      .then(() => setStatus("ok"))
      .catch((e) => {
        setStatus("error");
        setMessage(mapAuthError(e, "Verification failed"));
      });
  }, [token]);

  const onResend = async () => {
    if (!isAuthenticated) {
      toast({
        message: "Sign in first to resend the verification email.",
        variant: "warning",
      });
      return;
    }
    setResending(true);
    try {
      await resendVerification();
      toast({ message: "Verification email sent.", variant: "success" });
    } catch (e) {
      toast({ message: mapAuthError(e, "Could not resend email"), variant: "danger" });
    } finally {
      setResending(false);
    }
  };

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
          <Link href="/dashboard">Go to dashboard</Link>
        </Button>
      </div>
    );
  }

  if (status === "pending") {
    return (
      <div className="space-y-4 text-center">
        <Typography variant="h3">Verify your email</Typography>
        <p className="text-sm text-muted-foreground">
          {email
            ? `We sent a verification link to ${email}.`
            : "We sent a verification link to your email."}
        </p>
        <p className="text-sm text-muted-foreground">{message}</p>
        <div className="flex flex-col gap-2 sm:flex-row sm:justify-center">
          <Button loading={resending} onClick={() => void onResend()}>
            Resend verification email
          </Button>
          <Button asChild variant="outline">
            <Link href="/login">Sign in</Link>
          </Button>
        </div>
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
