package service

import (
	"context"
	"strings"
	"unicode"

	"duekeep/internal/model"
)

// KindStore — справочник типов. CountItems до Sprint 3 всегда 0.
type KindStore interface {
	List(ctx context.Context) ([]model.Kind, error)
	ByID(ctx context.Context, id string) (model.Kind, error)
	Create(ctx context.Context, k model.Kind) (model.Kind, error)
	Update(ctx context.Context, k model.Kind) (model.Kind, error)
	Delete(ctx context.Context, id string) error
	CountItems(ctx context.Context, id string) (int, error)
}

// Kind — CRUD item_kinds и валидация attr_schema.
type Kind struct {
	store KindStore
}

// NewKind собирает сервис типов.
func NewKind(store KindStore) *Kind {
	return &Kind{store: store}
}

// List все виды.
func (s *Kind) List(ctx context.Context) ([]model.Kind, error) {
	return s.store.List(ctx)
}

// Create проверяет slug/schema и пишет строку.
func (s *Kind) Create(ctx context.Context, in model.Kind) (model.Kind, error) {
	k, err := normalizeKind(in)
	if err != nil {
		return model.Kind{}, err
	}
	return s.store.Create(ctx, k)
}

// Patch частичное обновление. Пустой patch не ошибка.
func (s *Kind) Patch(ctx context.Context, id string, p model.KindPatch) (model.Kind, error) {
	cur, err := s.store.ByID(ctx, id)
	if err != nil {
		return model.Kind{}, err
	}
	if p.Slug != nil {
		cur.Slug = *p.Slug
	}
	if p.Name != nil {
		cur.Name = *p.Name
	}
	if p.Color != nil {
		cur.Color = *p.Color
	}
	if p.AttrSchema != nil {
		cur.AttrSchema = *p.AttrSchema
	}
	k, err := normalizeKind(cur)
	if err != nil {
		return model.Kind{}, err
	}
	k.ID = id
	return s.store.Update(ctx, k)
}

// Delete запрещён, если есть items (в Sprint 2 count всегда 0).
func (s *Kind) Delete(ctx context.Context, id string) error {
	if _, err := s.store.ByID(ctx, id); err != nil {
		return err
	}
	n, err := s.store.CountItems(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return model.ErrConflict
	}
	return s.store.Delete(ctx, id)
}

func normalizeKind(in model.Kind) (model.Kind, error) {
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	in.Color = strings.TrimSpace(in.Color)
	if in.Slug == "" || !validSlug(in.Slug) {
		return model.Kind{}, model.Validation("invalid slug", map[string]any{"slug": "lowercase [a-z0-9_-]"})
	}
	if in.Name == "" {
		return model.Kind{}, model.Validation(msgInvalidName, map[string]any{fieldName: detailRequired})
	}
	if in.Color == "" {
		return model.Kind{}, model.Validation("invalid color", map[string]any{"color": detailRequired})
	}
	if in.AttrSchema == nil {
		in.AttrSchema = []model.AttrField{}
	}
	if err := ValidateAttrSchema(in.AttrSchema); err != nil {
		return model.Kind{}, err
	}
	return in, nil
}

// ValidateAttrSchema — массив описателей, type ∈ string|number|boolean, key уникален.
func ValidateAttrSchema(fields []model.AttrField) error {
	seen := make(map[string]struct{}, len(fields))
	for i, f := range fields {
		key := strings.TrimSpace(f.Key)
		label := strings.TrimSpace(f.Label)
		if key == "" || label == "" {
			return model.Validation("invalid attr_schema", map[string]any{"index": i, "field": "key/label"})
		}
		if _, dup := seen[key]; dup {
			return model.Validation("duplicate attr key", map[string]any{"key": key})
		}
		seen[key] = struct{}{}
		switch f.Type {
		case model.AttrString, model.AttrNumber, model.AttrBoolean:
		default:
			return model.Validation("invalid attr type", map[string]any{"key": key, "type": f.Type})
		}
		fields[i].Key = key
		fields[i].Label = label
	}
	return nil
}

func validSlug(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9' && i > 0:
		case (r == '_' || r == '-') && i > 0:
		default:
			if unicode.IsUpper(r) {
				return false
			}
			return false
		}
	}
	return true
}
