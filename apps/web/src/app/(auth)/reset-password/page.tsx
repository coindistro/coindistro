"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  AlertDescription,
  Button,
  Field,
  PasswordInput,
  Typography,
} from "@coindistro/cds";
import { resetPassword } from "@/features/authentication/api";
import { resetSchema } from "@/features/authentication/schemas";
import { ApiError } from "@/lib/api/types";
import { z } from "zod";

type Values = z.infer<typeof resetSchema>;

export default function ResetPasswordPage() {
  const params = useSearchParams();
  const token = params.get("token") || "";
  const router = useRouter();
  const [error, setError] = useState<string | null>(
    token ? null : "Missing reset token.",
  );

  const form = useForm<Values>({
    resolver: zodResolver(resetSchema),
    defaultValues: { password: "", confirmPassword: "" },
  });

  const onSubmit = form.handleSubmit(async (values) => {
    setError(null);
    try {
      await resetPassword(token, values.password);
      router.replace("/login?reset=1");
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Reset failed");
    }
  });

  return (
    <div className="space-y-6">
      <div>
        <Typography variant="h3">Set a new password</Typography>
        <p className="mt-1 text-sm text-muted-foreground">Choose a strong password</p>
      </div>
      {error ? (
        <Alert variant="danger">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <form className="space-y-4" onSubmit={onSubmit}>
        <Field
          label="New password"
          htmlFor="password"
          error={form.formState.errors.password?.message}
          required
        >
          <PasswordInput id="password" {...form.register("password")} />
        </Field>
        <Field
          label="Confirm password"
          htmlFor="confirmPassword"
          error={form.formState.errors.confirmPassword?.message}
          required
        >
          <PasswordInput id="confirmPassword" {...form.register("confirmPassword")} />
        </Field>
        <Button
          type="submit"
          className="w-full"
          loading={form.formState.isSubmitting}
          disabled={!token}
        >
          Update password
        </Button>
      </form>
      <p className="text-center text-sm">
        <Link href="/login" className="text-primary hover:underline">
          Back to sign in
        </Link>
      </p>
    </div>
  );
}
