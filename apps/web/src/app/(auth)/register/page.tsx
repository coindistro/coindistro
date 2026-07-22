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
import {
  registerSchema,
  type RegisterValues,
} from "@/features/authentication/schemas";
import { ApiError, postLoginPath } from "@/lib/api/types";
import { useToast } from "@/features/shared/providers/toast-provider";

export default function RegisterPage() {
  const { register: registerUser } = useAuth();
  const router = useRouter();
  const params = useSearchParams();
  const { toast } = useToast();
  const [error, setError] = useState<string | null>(null);

  const form = useForm<RegisterValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      username: "",
      display_name: "",
      password: "",
      confirmPassword: "",
      referral_code: params.get("ref") || params.get("invite") || "",
    },
  });

  const onSubmit = form.handleSubmit(async (values) => {
    setError(null);
    try {
      const user = await registerUser({
        email: values.email,
        password: values.password,
        username: values.username || undefined,
        display_name: values.display_name || undefined,
        referral_code: values.referral_code,
      });
      toast({
        message: "Account created successfully. Welcome to Coindistro!",
        variant: "success",
      });
      router.replace(postLoginPath(user.roles));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Registration failed");
    }
  });

  return (
    <div className="space-y-6">
      <div>
        <Typography variant="h3">Create your account</Typography>
        <p className="mt-1 text-sm text-muted-foreground">
          Invite / referral code required for registration
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
        <Field label="Username" htmlFor="username" error={form.formState.errors.username?.message}>
          <Input id="username" autoComplete="username" {...form.register("username")} />
        </Field>
        <Field
          label="Referral code"
          htmlFor="referral_code"
          error={form.formState.errors.referral_code?.message}
          required
        >
          <Input id="referral_code" {...form.register("referral_code")} />
        </Field>
        <Field
          label="Password"
          htmlFor="password"
          error={form.formState.errors.password?.message}
          required
        >
          <PasswordInput id="password" autoComplete="new-password" {...form.register("password")} />
        </Field>
        <Field
          label="Confirm password"
          htmlFor="confirmPassword"
          error={form.formState.errors.confirmPassword?.message}
          required
        >
          <PasswordInput
            id="confirmPassword"
            autoComplete="new-password"
            {...form.register("confirmPassword")}
          />
        </Field>
        <Button type="submit" className="w-full" loading={form.formState.isSubmitting}>
          Create account
        </Button>
      </form>

      <p className="text-center text-sm text-muted-foreground">
        Already have an account?{" "}
        <Link href="/login" className="font-medium text-primary hover:underline">
          Sign in
        </Link>
      </p>
    </div>
  );
}
