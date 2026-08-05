package config

import (
	"fmt"
	"os"
	"strings"
)

// PaystackCredentials are loaded exclusively from environment variables so
// switching merchant accounts (e.g. PurpleSoftHub → CoinDistro) requires only
// env changes — no code modifications.
//
// Required variables (all must be set for production gateway operation):
//
//	PAYSTACK_SECRET_KEY      — secret key (server-side only; initialize + verify)
//	PAYSTACK_PUBLIC_KEY      — public key
//	PAYSTACK_CALLBACK_URL    — browser return URL after checkout
//	PAYSTACK_WEBHOOK_SECRET  — HMAC secret for x-paystack-signature verification
type PaystackCredentials struct {
	SecretKey     string
	PublicKey     string
	CallbackURL   string
	WebhookSecret string
}

// LoadPaystackCredentials reads Paystack credentials from the environment.
func LoadPaystackCredentials() PaystackCredentials {
	return PaystackCredentials{
		SecretKey:     envTrim("PAYSTACK_SECRET_KEY"),
		PublicKey:     envTrim("PAYSTACK_PUBLIC_KEY"),
		CallbackURL:   envTrim("PAYSTACK_CALLBACK_URL"),
		WebhookSecret: envTrim("PAYSTACK_WEBHOOK_SECRET"),
	}
}

// SignatureSecret returns the secret used to verify Paystack webhooks.
// Prefer PAYSTACK_WEBHOOK_SECRET; fall back to PAYSTACK_SECRET_KEY only when
// webhook secret is empty (should not happen after ValidateRequired).
func (c PaystackCredentials) SignatureSecret() string {
	if c.WebhookSecret != "" {
		return c.WebhookSecret
	}
	return c.SecretKey
}

// IsConfigured reports whether the secret key is present (minimum for init/verify).
func (c PaystackCredentials) IsConfigured() bool {
	return c.SecretKey != ""
}

// MissingRequired lists required env vars that are empty.
func (c PaystackCredentials) MissingRequired() []string {
	var missing []string
	if c.SecretKey == "" {
		missing = append(missing, "PAYSTACK_SECRET_KEY")
	}
	if c.PublicKey == "" {
		missing = append(missing, "PAYSTACK_PUBLIC_KEY")
	}
	if c.CallbackURL == "" {
		missing = append(missing, "PAYSTACK_CALLBACK_URL")
	}
	if c.WebhookSecret == "" {
		missing = append(missing, "PAYSTACK_WEBHOOK_SECRET")
	}
	return missing
}

// ValidateRequired fails when any required Paystack env var is missing.
// This prevents the API from serving EARNINGS_GATEWAY_NOT_CONFIGURED at runtime.
func (c PaystackCredentials) ValidateRequired() error {
	missing := c.MissingRequired()
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf(
		"Paystack gateway is not configured: missing %s. Set these on the backend host (Render Dashboard → Environment). Secrets must never be hardcoded",
		strings.Join(missing, ", "),
	)
}

// PresenceStatus returns human-readable Present/Missing for each required key
// without leaking secret values.
func (c PaystackCredentials) PresenceStatus() map[string]string {
	status := func(v string) string {
		if strings.TrimSpace(v) == "" {
			return "Missing"
		}
		return "Present"
	}
	return map[string]string{
		"PAYSTACK_SECRET_KEY":     status(c.SecretKey),
		"PAYSTACK_PUBLIC_KEY":     status(c.PublicKey),
		"PAYSTACK_CALLBACK_URL":   status(c.CallbackURL),
		"PAYSTACK_WEBHOOK_SECRET": status(c.WebhookSecret),
	}
}

func envTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
