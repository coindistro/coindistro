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
  Input,
  PasswordInput,
  Typography,
} from "@coindistro/cds";
import { useAuth } from "@/features/authentication/auth-provider";
import { loginSchema, type LoginValues } from "@/features/authentication/schemas";
import { ApiError, postLoginPath } from "@/lib/api/types";
import { useToast } from "@/features/shared/providers/toast-provider";

export default function LoginPage() {
  const { login } = useAuth();
  const router = useRouter();
  const params = useSearchParams();
  const { toast } = useToast();
  const [error, setError] = useState<string | null>(
    params.get("reason") === "session_expired"
      ? "Your session expired. Please sign in again."
      : null,
  );

  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  const onSubmit = form.handleSubmit(async (values) => {
    setError(null);
    try {
      const user = await login(values.email, values.password);
      toast({
        message: `Welcome back${user.display_name ? `, ${user.display_name}` : ""}`,
        variant: "success",
      });
      router.replace(postLoginPath(user.roles, params.get("next")));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Login failed");
    }
  });

  return (
    <div className="space-y-6">
      <div>
        <Typography variant="h3">Welcome back</Typography>
        <p className="mt-1 text-sm text-muted-foreground">
          Sign in to your Coindistro account
        </p>
      </div>

      {error ? (
        <Alert variant="danger">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <form className="space-y-4" onSubmit={onSubmit} noValidate>
        <Field label="Email" htmlFor="email" error={form.formState.errors.email?.message} required>
          <Input id="email" type="email" autoComplete="email" {...form.register("email")} />
        </Field>
        <Field
          label="Password"
          htmlFor="password"
          error={form.formState.errors.password?.message}
          required
        >
          <PasswordInput id="password" autoComplete="current-password" {...form.register("password")} />
        </Field>
        <div className="flex justify-end">
          <Link href="/forgot-password" className="text-sm text-primary hover:underline">
            Forgot password?
          </Link>
        </div>
        <Button type="submit" className="w-full" loading={form.formState.isSubmitting}>
          Sign in
        </Button>
      </form>

      <p className="text-center text-sm text-muted-foreground">
        New to Coindistro?{" "}
        <Link href="/register" className="font-medium text-primary hover:underline">
          Create an account
        </Link>
      </p>
    </div>
  );
}
