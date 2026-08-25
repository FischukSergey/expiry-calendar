package service_test

import (
	"testing"
	"time"

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
	if id.Role != string(model.RoleViewer) {
		t.Fatalf("role: %s", id.Role)
	}
	if id.UserID == "" {
		t.Fatal("empty sub")
	}
	if _, err := service.ParseAccess([]byte("wrong"), pair.AccessToken); err == nil {
		t.Fatal("want unauthorized for wrong secret")
	}
}
