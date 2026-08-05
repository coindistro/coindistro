package config

import (
	"os"
	"strings"
)

// PaystackCredentials are loaded exclusively from environment variables so
// switching merchant accounts (e.g. PurpleSoftHub → CoinDistro) requires only
// env changes — no code modifications.
//
// Required / supported variables:
//
//	PAYSTACK_SECRET_KEY      — secret key (server-side only; initialize + verify)
//	PAYSTACK_PUBLIC_KEY      — public key (optional; for clients that need it)
//	PAYSTACK_CALLBACK_URL    — redirect URL after checkout
//	PAYSTACK_WEBHOOK_SECRET  — HMAC secret for x-paystack-signature verification
type PaystackCredentials struct {
	SecretKey     string
	PublicKey     string
	CallbackURL   string
	WebhookSecret string
}

// LoadPaystackCredentials reads Paystack credentials from the environment.
// Empty fields mean the gateway is not fully configured for that capability.
func LoadPaystackCredentials() PaystackCredentials {
	return PaystackCredentials{
		SecretKey:     envTrim("PAYSTACK_SECRET_KEY"),
		PublicKey:     envTrim("PAYSTACK_PUBLIC_KEY"),
		CallbackURL:   envTrim("PAYSTACK_CALLBACK_URL"),
		WebhookSecret: envTrim("PAYSTACK_WEBHOOK_SECRET"),
	}
}

// SignatureSecret returns the secret used to verify Paystack webhooks.
// Prefer PAYSTACK_WEBHOOK_SECRET; fall back to PAYSTACK_SECRET_KEY (Paystack's default).
func (c PaystackCredentials) SignatureSecret() string {
	if c.WebhookSecret != "" {
		return c.WebhookSecret
	}
	return c.SecretKey
}

// IsConfigured reports whether the secret key is present (required for init/verify).
func (c PaystackCredentials) IsConfigured() bool {
	return c.SecretKey != ""
}

func envTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
