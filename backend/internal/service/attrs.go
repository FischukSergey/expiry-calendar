package service

import (
	"duekeep/internal/model"
)

// ValidateAttrs проверяет ключи и типы attrs против схемы. Лишние ключи — 422.
func ValidateAttrs(schema []model.AttrField, attrs map[string]any) error {
	if attrs == nil {
		attrs = map[string]any{}
	}
	allowed := make(map[string]model.AttrField, len(schema))
	for _, f := range schema {
		allowed[f.Key] = f
	}
	for key, val := range attrs {
		f, ok := allowed[key]
		if !ok {
			return model.Validation("unknown attr", map[string]any{fieldKey: key})
		}
		if !attrValueOK(f.Type, val) {
			return model.Validation("invalid attr type", map[string]any{fieldKey: key, fieldWant: f.Type})
		}
	}
	for _, f := range schema {
		if !f.Required {
			continue
		}
		if _, ok := attrs[f.Key]; !ok {
			return model.Validation("missing attr", map[string]any{fieldKey: f.Key})
		}
	}
	return nil
}

func attrValueOK(typ string, val any) bool {
	switch typ {
	case model.AttrString:
		_, ok := val.(string)
		return ok
	case model.AttrBoolean:
		_, ok := val.(bool)
		return ok
	case model.AttrNumber:
		switch val.(type) {
		case int, int32, int64, float32, float64:
			return true
		default:
			return false
		}
	default:
		return false
	}
}
