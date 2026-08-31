package service_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"duekeep/internal/clock"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

func TestHashRefreshStable(t *testing.T) {
	t.Parallel()
	raw := "opaque-refresh"
	a := service.HashRefresh(raw)
	b := service.HashRefresh(raw)
	if a != b {
		t.Fatal("hash must be stable")
	}
	if a == raw {
		t.Fatal("hash must not equal raw token")
	}
	if len(a) != 64 {
		t.Fatalf("sha256 hex length %d", len(a))
	}
	if service.HashRefresh("other") == a {
		t.Fatal("different raw must differ")
	}
}

func TestParseAccessClaims(t *testing.T) {
	t.Parallel()
	secret := []byte("unit-test-secret")
	svc := service.NewAuth(newMemUsers(), newMemRefresh(), nopTx, clock.Real{}, service.AuthConfig{
		Secret:     secret,
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 336 * time.Hour,
		BcryptCost: 4,
	})
	pair, err := svc.Register(t.Context(), "u@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	id, err := service.ParseAccess(secret, pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if id.Role != string(model.RoleAdmin) {
		t.Fatalf("role: %s", id.Role)
	}
	if id.UserID == "" {
		t.Fatal("empty sub")
	}

	parsed, _, err := jwt.NewParser().ParseUnverified(pair.AccessToken, jwt.MapClaims{})
	if err != nil {
		t.Fatal(err)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("claims type")
	}
	iss, err := claims.GetIssuer()
	if err != nil || iss != model.JWTIssuer {
		t.Fatalf("iss %q %v", iss, err)
	}
	sub, err := claims.GetSubject()
	if err != nil || sub != id.UserID {
		t.Fatalf("sub %q %v", sub, err)
	}
	exp, err := claims.GetExpirationTime()
	if err != nil || exp.Before(time.Now()) {
		t.Fatalf("exp %v %v", exp, err)
	}
	iat, err := claims.GetIssuedAt()
	if err != nil || iat.IsZero() {
		t.Fatalf("iat %v %v", iat, err)
	}
	if claims["role"] != string(model.RoleAdmin) {
		t.Fatalf("role claim %v", claims["role"])
	}
	if _, ok := claims["org_id"]; ok {
		t.Fatal("org_id must not be in access")
	}

	if _, err := service.ParseAccess([]byte("wrong"), pair.AccessToken); err == nil {
		t.Fatal("want unauthorized for wrong secret")
	}
}
