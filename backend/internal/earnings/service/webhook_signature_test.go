package service

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"testing"
)

func TestPaystackSignatureUsesWebhookSecret(t *testing.T) {
	payload := []byte(`{"event":"charge.success","data":{"reference":"ref-1"}}`)
	secret := "webhook-secret"
	svc := &Service{cfg: Config{PaystackWebhookSecret: secret}}

	mac := hmac.New(sha512.New, []byte(secret))
	_, _ = mac.Write(payload)
	valid := hex.EncodeToString(mac.Sum(nil))

	if !svc.verifyPaystackSignature(payload, valid) {
		t.Fatal("expected HMAC signature to verify")
	}
	plain := sha512.Sum512(payload)
	if svc.verifyPaystackSignature(payload, hex.EncodeToString(plain[:])) {
		t.Fatal("plain SHA-512 must not verify as a webhook signature")
	}
}
