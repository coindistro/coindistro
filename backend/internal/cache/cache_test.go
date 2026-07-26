package cache

import (
	"errors"
	"testing"

	cfg "github.com/coindistro/backend/internal/config"
)

func TestClassifyRedisError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "authentication", err: errors.New("WRONGPASS invalid username-password pair"), want: "authentication_failed"},
		{name: "timeout", err: errors.New("dial tcp: i/o timeout"), want: "timeout"},
		{name: "tls handshake", err: errors.New("tls: handshake failure"), want: "tls_handshake_failed"},
		{name: "connection refused", err: errors.New("connect: connection refused"), want: "connection_refused"},
		{name: "dns lookup", err: errors.New("lookup example.upstash.io: no such host"), want: "dns_lookup_failed"},
		{name: "eof", err: errors.New("EOF"), want: "eof"},
		{name: "generic", err: errors.New("something went wrong"), want: "unreachable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyRedisError(tt.err); got != tt.want {
				t.Fatalf("classifyRedisError(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}

func TestDiagnosticSummary(t *testing.T) {
	cache := &Cache{status: "healthy", config: cfg.RedisConfig{Host: "example.upstash.io", Port: 6379, TLSEnabled: true, Password: "secret"}}
	summary := cache.DiagnosticSummary()
	if summary["host"] != "example.upstash.io" {
		t.Fatalf("host = %v, want example.upstash.io", summary["host"])
	}
	if summary["password_set"] != true {
		t.Fatalf("password_set = %v, want true", summary["password_set"])
	}
}
