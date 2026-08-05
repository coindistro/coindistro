package config

import (
	"os"
	"testing"
)

func TestLoadPaystackCredentialsFromEnv(t *testing.T) {
	t.Setenv("PAYSTACK_SECRET_KEY", " sk_test_secret ")
	t.Setenv("PAYSTACK_PUBLIC_KEY", "pk_test_public")
	t.Setenv("PAYSTACK_CALLBACK_URL", "https://app.example.com/earn")
	t.Setenv("PAYSTACK_WEBHOOK_SECRET", "whsec_test")

	// Ensure legacy names are ignored (credentials must come from PAYSTACK_* only).
	t.Setenv("COINDISTRO_PAYSTACK_SECRET_KEY", "legacy_secret")
	t.Setenv("COINDISTRO_PAYSTACK_PUBLIC_KEY", "legacy_public")

	creds := LoadPaystackCredentials()
	if creds.SecretKey != "sk_test_secret" {
		t.Fatalf("secret = %q, want sk_test_secret", creds.SecretKey)
	}
	if creds.PublicKey != "pk_test_public" {
		t.Fatalf("public = %q, want pk_test_public", creds.PublicKey)
	}
	if creds.CallbackURL != "https://app.example.com/earn" {
		t.Fatalf("callback = %q", creds.CallbackURL)
	}
	if creds.WebhookSecret != "whsec_test" {
		t.Fatalf("webhook secret = %q", creds.WebhookSecret)
	}
	if !creds.IsConfigured() {
		t.Fatal("expected configured when secret is set")
	}
	if got := creds.SignatureSecret(); got != "whsec_test" {
		t.Fatalf("signature secret = %q, want webhook secret", got)
	}
}

func TestSignatureSecretFallsBackToSecretKey(t *testing.T) {
	t.Setenv("PAYSTACK_SECRET_KEY", "sk_live")
	_ = os.Unsetenv("PAYSTACK_WEBHOOK_SECRET")
	// t.Setenv cannot unset; overwrite empty
	t.Setenv("PAYSTACK_WEBHOOK_SECRET", "")
	t.Setenv("PAYSTACK_PUBLIC_KEY", "")
	t.Setenv("PAYSTACK_CALLBACK_URL", "")

	creds := LoadPaystackCredentials()
	if got := creds.SignatureSecret(); got != "sk_live" {
		t.Fatalf("signature secret = %q, want sk_live", got)
	}
}

func TestLoadPaystackCredentialsEmpty(t *testing.T) {
	t.Setenv("PAYSTACK_SECRET_KEY", "")
	t.Setenv("PAYSTACK_PUBLIC_KEY", "")
	t.Setenv("PAYSTACK_CALLBACK_URL", "")
	t.Setenv("PAYSTACK_WEBHOOK_SECRET", "")

	creds := LoadPaystackCredentials()
	if creds.IsConfigured() {
		t.Fatal("expected not configured when secret empty")
	}
}
