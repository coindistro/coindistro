package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/auth"
	"github.com/coindistro/backend/internal/config"
)

func setupAuth() *auth.Auth {
	cfg := config.AuthConfig{
		AccessTokenSecret:  "test-access-secret-key",
		RefreshTokenSecret: "test-refresh-secret-key",
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    10080 * time.Minute,
		Issuer:             "coindistro-test",
	}
	logger, _ := zap.NewDevelopment()
	return auth.New(cfg, logger)
}

func TestGenerateAccessToken(t *testing.T) {
	a := setupAuth()
	token, err := a.GenerateAccessToken("user-123", "test@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	a := setupAuth()
	token, err := a.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateAccessToken(t *testing.T) {
	a := setupAuth()
	token, err := a.GenerateAccessToken("user-123", "test@example.com", []string{"user", "admin"})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := a.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected user_id 'user-123', got '%s'", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", claims.Email)
	}
	if len(claims.Roles) != 2 || claims.Roles[0] != "user" || claims.Roles[1] != "admin" {
		t.Errorf("expected roles [user admin], got %v", claims.Roles)
	}
	if claims.Issuer != "coindistro-test" {
		t.Errorf("expected issuer 'coindistro-test', got '%s'", claims.Issuer)
	}
}

func TestValidateRefreshToken(t *testing.T) {
	a := setupAuth()
	token, err := a.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := a.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("expected user_id 'user-123', got '%s'", claims.UserID)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	a := setupAuth()
	_, err := a.ValidateAccessToken("invalid-token-string")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateExpiredToken(t *testing.T) {
	// Create auth with expired token
	cfg := config.AuthConfig{
		AccessTokenSecret:  "test-access-secret-key",
		RefreshTokenSecret: "test-refresh-secret-key",
		AccessTokenTTL:     -1 * time.Minute, // Already expired
		RefreshTokenTTL:    10080 * time.Minute,
		Issuer:             "coindistro-test",
	}
	logger, _ := zap.NewDevelopment()
	a := auth.New(cfg, logger)

	token, err := a.GenerateAccessToken("user-123", "test@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = a.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestTokenPair(t *testing.T) {
	a := setupAuth()
	pair, err := a.GenerateTokenPair("user-123", "test@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("expected token type 'Bearer', got '%s'", pair.TokenType)
	}
	if pair.ExpiresIn <= 0 {
		t.Errorf("expected positive expires_in, got %d", pair.ExpiresIn)
	}
}

func TestHashPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == password {
		t.Fatal("hash should not equal plain text password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if err := auth.VerifyPassword(password, hash); err != nil {
		t.Errorf("expected password verification to succeed: %v", err)
	}

	if err := auth.VerifyPassword("WrongPassword123!", hash); err == nil {
		t.Error("expected password verification to fail for wrong password")
	}
}

func TestSetUserContext(t *testing.T) {
	ctx := auth.SetUserContext(nil, "user-123", "test@example.com", []string{"user"})
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	userID, ok := auth.GetUserID(ctx)
	if !ok || userID != "user-123" {
		t.Errorf("expected user_id 'user-123', got '%s'", userID)
	}

	email, ok := auth.GetEmail(ctx)
	if !ok || email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", email)
	}

	roles, ok := auth.GetRoles(ctx)
	if !ok || len(roles) != 1 || roles[0] != "user" {
		t.Errorf("expected roles [user], got %v", roles)
	}
}

func TestTokenSigningMethod(t *testing.T) {
	a := setupAuth()
	token, err := a.GenerateAccessToken("user-123", "test@example.com", []string{"user"})
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Parse without validation to check signing method
	parsed, _, _ := jwt.NewParser().ParseUnverified(token, &auth.Claims{})
	if parsed.Method.Alg() != "HS256" {
		t.Errorf("expected signing method HS256, got %s", parsed.Method.Alg())
	}
}
