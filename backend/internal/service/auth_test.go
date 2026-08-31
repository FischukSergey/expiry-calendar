package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"duekeep/internal/clock"
	"duekeep/internal/model"
	"duekeep/internal/seed"
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

type memCatWriter struct {
	rows []model.Category
}

func (m *memCatWriter) Create(_ context.Context, c model.Category) (model.Category, error) {
	c.ID = uuid.NewString()
	c.Children = []model.Category{}
	m.rows = append(m.rows, c)
	return c, nil
}

func TestRegisterCopiesDefaultCategories(t *testing.T) {
	t.Parallel()
	want := seed.DefaultCategories()
	cats := &memCatWriter{}
	svc := testAuth(t)
	svc.SetCategoryDefaults(cats)

	pair, err := svc.Register(t.Context(), "cats@duekeep.local", "secret12", "")
	if err != nil {
		t.Fatal(err)
	}
	id, err := service.ParseAccess([]byte("unit-test-secret"), pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats.rows) != len(want) {
		t.Fatalf("cats %d want %d", len(cats.rows), len(want))
	}
	for i, c := range cats.rows {
		if c.OwnerID != id.UserID {
			t.Fatalf("owner %s want %s", c.OwnerID, id.UserID)
		}
		if c.Name != want[i].Name || c.SortOrder != want[i].SortOrder {
			t.Fatalf("row %d %+v want %+v", i, c, want[i])
		}
		if want[i].ParentIdx < 0 {
			if c.ParentID != nil && *c.ParentID != "" {
				t.Fatalf("root %s has parent", c.Name)
			}
		} else {
			parent := cats.rows[want[i].ParentIdx]
			if c.ParentID == nil || *c.ParentID != parent.ID {
				t.Fatalf("%s parent %v want %s", c.Name, c.ParentID, parent.ID)
			}
		}
	}

	other := &memCatWriter{}
	second := testAuth(t)
	second.SetCategoryDefaults(other)
	if _, err := second.Register(t.Context(), "other@duekeep.local", "secret12", ""); err != nil {
		t.Fatal(err)
	}
	if len(other.rows) != len(want) {
		t.Fatalf("second cats %d", len(other.rows))
	}
	if other.rows[0].ID == cats.rows[0].ID {
		t.Fatal("category ids must be per-user")
	}
	if other.rows[0].OwnerID == cats.rows[0].OwnerID {
		t.Fatal("owners must differ")
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
