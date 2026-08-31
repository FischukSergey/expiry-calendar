package service_test

import (
	"errors"
	"testing"
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

func testAuth(t *testing.T) *service.Auth {
	t.Helper()
	return service.NewAuth(newMemUsers(), newMemRefresh(), nopTx, clock.Real{}, service.AuthConfig{
		Secret:     []byte("unit-test-secret"),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 336 * time.Hour,
		BcryptCost: 4,
	})
}

func TestRegisterCreatesAdmin(t *testing.T) {
	t.Parallel()
	svc := testAuth(t)
	pair, err := svc.Register(t.Context(), "new@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	id, err := service.ParseAccess([]byte("unit-test-secret"), pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if id.Role != string(model.RoleAdmin) {
		t.Fatalf("role: %s", id.Role)
	}
	if pair.TokenType != model.TokenTypeBearer || pair.ExpiresIn != 900 {
		t.Fatalf("pair: %+v", pair)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	t.Parallel()
	svc := testAuth(t)
	if _, err := svc.Register(t.Context(), "dup@duekeep.local", "secret12", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(t.Context(), "dup@duekeep.local", "secret12", ""); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("got %v", err)
	}
}

func TestLoginBadPassword(t *testing.T) {
	t.Parallel()
	svc := testAuth(t)
	if _, err := svc.Register(t.Context(), "login@duekeep.local", "secret12", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Login(t.Context(), "login@duekeep.local", "wrongpass", ""); !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}

func TestRefreshRotationAndReuse(t *testing.T) {
	t.Parallel()
	svc := testAuth(t)
	first, err := svc.Register(t.Context(), "rot@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Refresh(t.Context(), first.RefreshToken, "")
	if err != nil {
		t.Fatal(err)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh must rotate")
	}
	if _, err := svc.Refresh(t.Context(), first.RefreshToken, ""); !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("old refresh: %v", err)
	}
	if _, err := svc.Refresh(t.Context(), second.RefreshToken, ""); !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("family after reuse: %v", err)
	}
}

func TestUnknownRefreshNoRevoke(t *testing.T) {
	t.Parallel()
	svc := testAuth(t)
	pair, err := svc.Register(t.Context(), "unk@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Refresh(t.Context(), "not-a-real-token", ""); !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
	if _, err := svc.Refresh(t.Context(), pair.RefreshToken, ""); err != nil {
		t.Fatalf("legitimate refresh after garbage: %v", err)
	}
}

func TestLogoutThenRefresh(t *testing.T) {
	t.Parallel()
	svc := testAuth(t)
	pair, err := svc.Register(t.Context(), "out@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(t.Context(), pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(t.Context(), pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Refresh(t.Context(), pair.RefreshToken, ""); !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}

func TestLogoutAll(t *testing.T) {
	t.Parallel()
	svc := testAuth(t)
	pair, err := svc.Register(t.Context(), "all@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	id, err := service.ParseAccess([]byte("unit-test-secret"), pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.LogoutAll(t.Context(), id.UserID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Refresh(t.Context(), pair.RefreshToken, ""); !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}

func TestLoginRefreshClaimsNoOrgID(t *testing.T) {
	t.Parallel()
	secret := []byte("unit-test-secret")
	svc := testAuth(t)
	reg, err := svc.Register(t.Context(), "lr@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	login, err := svc.Login(t.Context(), "lr@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	ref, err := svc.Refresh(t.Context(), login.RefreshToken, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{reg.AccessToken, login.AccessToken, ref.AccessToken} {
		id, err := service.ParseAccess(secret, raw)
		if err != nil {
			t.Fatal(err)
		}
		if id.Role != string(model.RoleAdmin) || id.UserID == "" {
			t.Fatalf("%+v", id)
		}
	}
}

func TestMe(t *testing.T) {
	t.Parallel()
	svc := testAuth(t)
	pair, err := svc.Register(t.Context(), "me@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	id, err := service.ParseAccess([]byte("unit-test-secret"), pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	me, err := svc.Me(t.Context(), id.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if me.Email != "me@duekeep.local" || me.Role != model.RoleAdmin {
		t.Fatalf("%+v", me)
	}
}
