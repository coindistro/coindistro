"use client";

import { useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Alert,
  AlertDescription,
  Button,
  Field,
  Input,
  Typography,
} from "@coindistro/cds";
import { forgotPassword } from "@/features/authentication/api";
import { forgotSchema } from "@/features/authentication/schemas";
import { ApiError } from "@/lib/api/types";
import { z } from "zod";

type Values = z.infer<typeof forgotSchema>;

export default function ForgotPasswordPage() {
  const [done, setDone] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const form = useForm<Values>({
    resolver: zodResolver(forgotSchema),
    defaultValues: { email: "" },
  });

  const onSubmit = form.handleSubmit(async (values) => {
    setError(null);
    try {
      await forgotPassword(values.email);
      setDone(true);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Request failed");
    }
  });

  if (done) {
    return (
      <div className="space-y-4 text-center">
        <Typography variant="h3">Check your email</Typography>
        <p className="text-sm text-muted-foreground">
          If an account exists for that address, we sent reset instructions.
        </p>
        <Link href="/login" className="text-sm text-primary hover:underline">
          Back to sign in
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <Typography variant="h3">Forgot password</Typography>
        <p className="mt-1 text-sm text-muted-foreground">
          We&apos;ll email you a reset link
        </p>
      </div>
      {error ? (
        <Alert variant="danger">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <form className="space-y-4" onSubmit={onSubmit}>
        <Field label="Email" htmlFor="email" error={form.formState.errors.email?.message} required>
          <Input id="email" type="email" {...form.register("email")} />
        </Field>
        <Button type="submit" className="w-full" loading={form.formState.isSubmitting}>
          Send reset link
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
