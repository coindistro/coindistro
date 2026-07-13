package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/coindistro/backend/internal/config"
	apperrors "github.com/coindistro/backend/internal/errors"
)

// TokenType represents the type of JWT token.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the custom JWT claims.
type Claims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// Auth handles authentication operations.
type Auth struct {
	config config.AuthConfig
	logger *zap.Logger
}

// New creates a new Auth instance.
func New(cfg config.AuthConfig, logger *zap.Logger) *Auth {
	return &Auth{
		config: cfg,
		logger: logger,
	}
}

// GenerateAccessToken generates a new access token for a user.
func (a *Auth) GenerateAccessToken(userID string, email string, roles []string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.config.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    a.config.Issuer,
			Subject:   userID,
			ID:        fmt.Sprintf("%s-%d", userID, now.UnixNano()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(a.config.AccessTokenSecret))
	if err != nil {
		a.logger.Error("failed to sign access token", zap.Error(err))
		return "", apperrors.Wrap(err, "TOKEN_GENERATION_FAILED", "Failed to generate access token", 500)
	}

	return signedToken, nil
}

// GenerateRefreshToken generates a new refresh token for a user.
func (a *Auth) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(a.config.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    a.config.Issuer,
			Subject:   userID,
			ID:        fmt.Sprintf("refresh-%s-%d", userID, now.UnixNano()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(a.config.RefreshTokenSecret))
	if err != nil {
		a.logger.Error("failed to sign refresh token", zap.Error(err))
		return "", apperrors.Wrap(err, "TOKEN_GENERATION_FAILED", "Failed to generate refresh token", 500)
	}

	return signedToken, nil
}

// ValidateAccessToken validates an access token and returns the claims.
func (a *Auth) ValidateAccessToken(tokenString string) (*Claims, error) {
	return a.validateToken(tokenString, a.config.AccessTokenSecret, AccessToken)
}

// ValidateRefreshToken validates a refresh token and returns the claims.
func (a *Auth) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return a.validateToken(tokenString, a.config.RefreshTokenSecret, RefreshToken)
}

func (a *Auth) validateToken(tokenString string, secret string, tokenType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		switch {
		case jwt.ErrTokenExpired.Error() == err.Error():
			return nil, apperrors.ErrTokenExpired
		case jwt.ErrTokenMalformed.Error() == err.Error():
			return nil, apperrors.ErrInvalidToken
		default:
			return nil, apperrors.ErrInvalidToken
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrInvalidToken
	}

	// Verify issuer
	if claims.Issuer != a.config.Issuer {
		return nil, apperrors.ErrInvalidToken
	}

	return claims, nil
}

// HashPassword hashes a plain text password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperrors.Wrap(err, "PASSWORD_HASH_FAILED", "Failed to hash password", 500)
	}
	return string(bytes), nil
}

// VerifyPassword compares a plain text password with a bcrypt hash.
func VerifyPassword(password string, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return apperrors.ErrUnauthorized
	}
	return nil
}

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// GenerateTokenPair generates both access and refresh tokens.
func (a *Auth) GenerateTokenPair(userID string, email string, roles []string) (*TokenPair, error) {
	accessToken, err := a.GenerateAccessToken(userID, email, roles)
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(a.config.AccessTokenTTL.Seconds()),
	}, nil
}

const (
	// ContextKeyUserID is the key for storing the user ID in context.
	ContextKeyUserID = "user_id"
	// ContextKeyEmail is the key for storing the email in context.
	ContextKeyEmail = "email"
	// ContextKeyRoles is the key for storing the roles in context.
	ContextKeyRoles = "roles"
)

// SetUserContext stores authenticated user information in the request context.
func SetUserContext(ctx context.Context, userID string, email string, roles []string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, ContextKeyEmail, email)
	ctx = context.WithValue(ctx, ContextKeyRoles, roles)
	return ctx
}

// GetUserID retrieves the user ID from the context.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(string)
	return userID, ok
}

// GetEmail retrieves the email from the context.
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(ContextKeyEmail).(string)
	return email, ok
}

// GetRoles retrieves the roles from the context.
func GetRoles(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(ContextKeyRoles).([]string)
	return roles, ok
}
