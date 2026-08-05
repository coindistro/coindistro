package config

import (
	"strings"
	"testing"
)

func TestLoadPaystackCredentialsFromEnv(t *testing.T) {
	t.Setenv("PAYSTACK_SECRET_KEY", " sk_test_secret ")
	t.Setenv("PAYSTACK_PUBLIC_KEY", "pk_test_public")
	t.Setenv("PAYSTACK_CALLBACK_URL", "https://coindistro-hazel.vercel.app/app/earn")
	t.Setenv("PAYSTACK_WEBHOOK_SECRET", "whsec_test")

	// Legacy names must not be used.
	t.Setenv("COINDISTRO_PAYSTACK_SECRET_KEY", "legacy_secret")
	t.Setenv("COINDISTRO_PAYSTACK_PUBLIC_KEY", "legacy_public")

	creds := LoadPaystackCredentials()
	if creds.SecretKey != "sk_test_secret" {
		t.Fatalf("secret = %q, want sk_test_secret", creds.SecretKey)
	}
	if creds.PublicKey != "pk_test_public" {
		t.Fatalf("public = %q, want pk_test_public", creds.PublicKey)
	}
	if creds.CallbackURL != "https://coindistro-hazel.vercel.app/app/earn" {
		t.Fatalf("callback = %q", creds.CallbackURL)
	}
	if creds.WebhookSecret != "whsec_test" {
		t.Fatalf("webhook secret = %q", creds.WebhookSecret)
	}
	if err := creds.ValidateRequired(); err != nil {
		t.Fatalf("ValidateRequired: %v", err)
	}
	if got := creds.SignatureSecret(); got != "whsec_test" {
		t.Fatalf("signature secret = %q, want webhook secret", got)
	}
	presence := creds.PresenceStatus()
	for _, key := range []string{
		"PAYSTACK_SECRET_KEY",
		"PAYSTACK_PUBLIC_KEY",
		"PAYSTACK_CALLBACK_URL",
		"PAYSTACK_WEBHOOK_SECRET",
	} {
		if presence[key] != "Present" {
			t.Fatalf("%s status = %q, want Present", key, presence[key])
		}
	}
}

func TestValidateRequiredFailsWhenMissing(t *testing.T) {
	t.Setenv("PAYSTACK_SECRET_KEY", "")
	t.Setenv("PAYSTACK_PUBLIC_KEY", "")
	t.Setenv("PAYSTACK_CALLBACK_URL", "")
	t.Setenv("PAYSTACK_WEBHOOK_SECRET", "")

	creds := LoadPaystackCredentials()
	err := creds.ValidateRequired()
	if err == nil {
		t.Fatal("expected validation error when keys missing")
	}
	if !strings.Contains(err.Error(), "PAYSTACK_SECRET_KEY") {
		t.Fatalf("error should list missing secret: %v", err)
	}
	presence := creds.PresenceStatus()
	if presence["PAYSTACK_SECRET_KEY"] != "Missing" {
		t.Fatalf("expected Missing, got %q", presence["PAYSTACK_SECRET_KEY"])
	}
}

func TestSignatureSecretFallsBackToSecretKey(t *testing.T) {
	t.Setenv("PAYSTACK_SECRET_KEY", "sk_live")
	t.Setenv("PAYSTACK_WEBHOOK_SECRET", "")
	t.Setenv("PAYSTACK_PUBLIC_KEY", "pk")
	t.Setenv("PAYSTACK_CALLBACK_URL", "https://example.com")

	creds := LoadPaystackCredentials()
	if got := creds.SignatureSecret(); got != "sk_live" {
		t.Fatalf("signature secret = %q, want sk_live", got)
	}
}
