package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"duekeep/internal/model"
)

const refreshRawBytes = 32

// accessClaims — HS256 payload. Access в БД не храним.
type accessClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// HashRefresh — SHA-256 hex сырого refresh. Один вход → один хеш.
func HashRefresh(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// NewRefreshRaw — непрозрачная строка (32 байта, raw URL base64).
func NewRefreshRaw() (string, error) {
	buf := make([]byte, refreshRawBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("refresh rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func issueAccess(secret []byte, user model.User, now time.Time, ttl time.Duration) (string, error) {
	claims := accessClaims{
		Role: string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    model.JWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign access: %w", err)
	}
	return signed, nil
}

// AccessIdentity — sub и role из валидного access.
type AccessIdentity struct {
	UserID string
	Role   string
}

// ParseAccess проверяет подпись HS256, iss и exp. Битый/чужой alg → ошибка.
func ParseAccess(secret []byte, token string) (AccessIdentity, error) {
	var claims accessClaims
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, model.ErrUnauthorized
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(model.JWTIssuer))
	if err != nil || !parsed.Valid || claims.Subject == "" || claims.Role == "" {
		return AccessIdentity{}, model.ErrUnauthorized
	}
	return AccessIdentity{UserID: claims.Subject, Role: claims.Role}, nil
}
