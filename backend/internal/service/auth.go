package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"duekeep/internal/clock"
	"duekeep/internal/model"
	"duekeep/internal/seed"
)

// UserStore — пользователи. Интерфейс объявлен у потребителя.
type UserStore interface {
	Create(ctx context.Context, email, passwordHash string, role model.Role) (model.User, error)
	ByEmail(ctx context.Context, email string) (model.User, error)
	ByID(ctx context.Context, id string) (model.User, error)
}

// RefreshStore — сессии refresh. Сырой токен store не видит.
type RefreshStore interface {
	Insert(ctx context.Context, rec model.RefreshSession) error
	ByHash(ctx context.Context, hash string) (model.RefreshSession, error)
	ByHashForUpdate(ctx context.Context, hash string) (model.RefreshSession, error)
	RevokeID(ctx context.Context, id string, at time.Time) error
	RevokeFamily(ctx context.Context, familyID string, at time.Time) error
	RevokeUser(ctx context.Context, userID string, at time.Time) error
}

// TxFunc оборачивает сценарии, где user + refresh должны писаться вместе.
type TxFunc func(ctx context.Context, fn func(context.Context) error) error

// AuthConfig — секреты и TTL из env. BcryptCost 0 → DefaultCost.
type AuthConfig struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	BcryptCost int
}

// CategoryWriter — INSERT категории (копия дерева при Register).
type CategoryWriter interface {
	Create(ctx context.Context, c model.Category) (model.Category, error)
}

// Auth — register/login/refresh/logout. Cookie vs body сюда не входят: только сырой refresh.
type Auth struct {
	users   UserStore
	tokens  RefreshStore
	tx      TxFunc
	clk     clock.Clock
	secret  []byte
	access  time.Duration
	refresh time.Duration
	bcrypt  int
	cats    CategoryWriter
}

// NewAuth собирает сервис. Secret не должен быть пустым (проверяет main).
func NewAuth(users UserStore, tokens RefreshStore, tx TxFunc, clk clock.Clock, cfg AuthConfig) *Auth {
	cost := cfg.BcryptCost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &Auth{
		users:   users,
		tokens:  tokens,
		tx:      tx,
		clk:     clk,
		secret:  cfg.Secret,
		access:  cfg.AccessTTL,
		refresh: cfg.RefreshTTL,
		bcrypt:  cost,
	}
}

// SetCategoryDefaults копирует дефолтное дерево категорий при Register. Без вызова — только user+tokens.
func (s *Auth) SetCategoryDefaults(w CategoryWriter) {
	s.cats = w
}

// ExpiresIn — секунды access для JSON expires_in.
func (s *Auth) ExpiresIn() int {
	return int(s.access / time.Second)
}

// Register создаёт admin и копию дефолтных категорий (без items). Дубликат email → ErrConflict.
func (s *Auth) Register(ctx context.Context, email, password, userAgent string) (model.TokenPair, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return model.TokenPair{}, err
	}
	if err := validatePassword(password); err != nil {
		return model.TokenPair{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcrypt)
	if err != nil {
		return model.TokenPair{}, err
	}

	var user model.User
	var pair model.TokenPair
	err = s.tx(ctx, func(ctx context.Context) error {
		created, cerr := s.users.Create(ctx, email, string(hash), model.RoleAdmin)
		if cerr != nil {
			return cerr
		}
		user = created
		if s.cats != nil {
			if cerr := copyDefaultCategories(ctx, user.ID, s.cats); cerr != nil {
				return cerr
			}
		}
		p, perr := s.issuePair(ctx, user, uuid.NewString(), userAgent)
		if perr != nil {
			return perr
		}
		pair = p
		return nil
	})
	if err != nil {
		return model.TokenPair{}, err
	}
	return pair, nil
}

// Login сверяет bcrypt. Неверный email или пароль → один ErrUnauthorized.
func (s *Auth) Login(ctx context.Context, email, password, userAgent string) (model.TokenPair, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return model.TokenPair{}, model.ErrUnauthorized
	}
	if password == "" {
		return model.TokenPair{}, model.ErrUnauthorized
	}
	user, err := s.users.ByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.TokenPair{}, model.ErrUnauthorized
		}
		return model.TokenPair{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return model.TokenPair{}, model.ErrUnauthorized
	}

	var pair model.TokenPair
	err = s.tx(ctx, func(ctx context.Context) error {
		p, perr := s.issuePair(ctx, user, uuid.NewString(), userAgent)
		if perr != nil {
			return perr
		}
		pair = p
		return nil
	})
	if err != nil {
		return model.TokenPair{}, err
	}
	return pair, nil
}

// Refresh крутит пару с тем же family_id. Неизвестный токен — 401 без revoke.
// Уже отозванный в семье — revoke всей family и 401.
func (s *Auth) Refresh(ctx context.Context, raw, userAgent string) (model.TokenPair, error) {
	if raw == "" {
		return model.TokenPair{}, model.ErrUnauthorized
	}
	hash := HashRefresh(raw)
	now := s.clk.Now()

	var pair model.TokenPair
	err := s.tx(ctx, func(ctx context.Context) error {
		rec, err := s.tokens.ByHashForUpdate(ctx, hash)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return model.ErrUnauthorized
			}
			return err
		}
		if rec.RevokedAt != nil {
			_ = s.tokens.RevokeFamily(ctx, rec.FamilyID, now)
			return model.ErrUnauthorized
		}
		if !rec.ExpiresAt.After(now) {
			return model.ErrUnauthorized
		}
		user, err := s.users.ByID(ctx, rec.UserID)
		if err != nil {
			return err
		}
		if err := s.tokens.RevokeID(ctx, rec.ID, now); err != nil {
			return err
		}
		p, err := s.issuePair(ctx, user, rec.FamilyID, userAgent)
		if err != nil {
			return err
		}
		pair = p
		return nil
	})
	if err != nil {
		return model.TokenPair{}, err
	}
	return pair, nil
}

// Logout отзывает refresh, если он известен и ещё жив. Неизвестный — не ошибка (204).
func (s *Auth) Logout(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	rec, err := s.tokens.ByHash(ctx, HashRefresh(raw))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil
		}
		return err
	}
	if rec.RevokedAt != nil {
		return nil
	}
	return s.tokens.RevokeID(ctx, rec.ID, s.clk.Now())
}

// LogoutAll отзывает все семьи пользователя. Access после этого жив до exp.
func (s *Auth) LogoutAll(ctx context.Context, userID string) error {
	return s.tokens.RevokeUser(ctx, userID, s.clk.Now())
}

// Me возвращает публичного пользователя по sub из access.
func (s *Auth) Me(ctx context.Context, userID string) (model.PublicUser, error) {
	user, err := s.users.ByID(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.PublicUser{}, model.ErrUnauthorized
		}
		return model.PublicUser{}, err
	}
	return user.Public(), nil
}

func (s *Auth) issuePair(ctx context.Context, user model.User, familyID, userAgent string) (model.TokenPair, error) {
	now := s.clk.Now()
	access, err := issueAccess(s.secret, user, now, s.access)
	if err != nil {
		return model.TokenPair{}, err
	}
	raw, err := NewRefreshRaw()
	if err != nil {
		return model.TokenPair{}, err
	}
	rec := model.RefreshSession{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		FamilyID:  familyID,
		TokenHash: HashRefresh(raw),
		ExpiresAt: now.Add(s.refresh),
		UserAgent: userAgent,
		CreatedAt: now,
	}
	if err := s.tokens.Insert(ctx, rec); err != nil {
		return model.TokenPair{}, err
	}
	return model.TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		TokenType:    model.TokenTypeBearer,
		ExpiresIn:    s.ExpiresIn(),
	}, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" || !strings.Contains(email, "@") {
		return "", model.Validation("invalid email", map[string]any{"email": detailRequired})
	}
	return email, nil
}

func validatePassword(password string) error {
	if len(password) < model.MinPasswordLen {
		return model.Validation("password too short", map[string]any{"password": "min 8"})
	}
	return nil
}

// copyDefaultCategories пишет шаблон дерева владельцу. Новые id, без items.
func copyDefaultCategories(ctx context.Context, ownerID string, w CategoryWriter) error {
	templates := seed.DefaultCategories()
	ids := make([]string, len(templates))
	for i, t := range templates {
		if t.ParentIdx >= i {
			return fmt.Errorf("default category %s: parent after child", t.Name)
		}
		var parent *string
		if t.ParentIdx >= 0 {
			p := ids[t.ParentIdx]
			parent = &p
		}
		created, err := w.Create(ctx, model.Category{
			OwnerID:   ownerID,
			ParentID:  parent,
			Name:      t.Name,
			SortOrder: t.SortOrder,
		})
		if err != nil {
			return fmt.Errorf("copy category %s: %w", t.Name, err)
		}
		ids[i] = created.ID
	}
	return nil
}
