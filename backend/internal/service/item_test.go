package service_test

import (
	"testing"
	"time"

	"duekeep/internal/model"
	"duekeep/internal/service"
)

func TestValidateAttrs(t *testing.T) {
	t.Parallel()
	const seats = "seats"
	schema := []model.AttrField{
		{Key: seats, Label: "Места", Type: model.AttrNumber, Required: true},
		{Key: "auto_renew", Label: "Авто", Type: model.AttrBoolean},
	}
	if err := service.ValidateAttrs(schema, map[string]any{seats: 5, "auto_renew": true}); err != nil {
		t.Fatal(err)
	}
	if err := service.ValidateAttrs(schema, map[string]any{seats: 5, "extra": "x"}); err == nil {
		t.Fatal("extra key")
	}
	if err := service.ValidateAttrs(schema, map[string]any{}); err == nil {
		t.Fatal("missing required")
	}
	if err := service.ValidateAttrs(nil, map[string]any{"x": 1}); err == nil {
		t.Fatal("empty schema extra")
	}
	if err := service.ValidateAttrs(nil, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if err := service.ValidateAttrs(schema, map[string]any{seats: "5"}); err == nil {
		t.Fatal("wrong type")
	}
}

func TestStatusAtWrite(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		expire time.Time
		req    string
		want   string
	}{
		{"expired", today.AddDate(0, 0, -1), "", model.StatusExpired},
		{"today", today, "", model.StatusExpiring},
		{"edge", today.AddDate(0, 0, 30), "", model.StatusExpiring},
		{"active", today.AddDate(0, 0, 31), "", model.StatusActive},
		{"keep cancelled", today.AddDate(0, 0, -1), model.StatusCancelled, model.StatusCancelled},
		{"keep archived", today.AddDate(0, 0, 90), model.StatusArchived, model.StatusArchived},
		{"ignore client active", today.AddDate(0, 0, -1), model.StatusActive, model.StatusExpired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := service.StatusAtWrite(today, tc.expire, 30, tc.req)
			if got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestNormalizePage(t *testing.T) {
	t.Parallel()
	p, err := service.NormalizePage(0, 0)
	if err != nil || p.Page != 1 || p.PerPage != 20 {
		t.Fatalf("defaults %+v %v", p, err)
	}
	p, err = service.NormalizePage(2, 200)
	if err != nil || p.PerPage != 100 || p.Offset() != 100 {
		t.Fatalf("clamp %+v %v", p, err)
	}
	if _, err := service.NormalizePage(-1, 10); err == nil {
		t.Fatal("neg page")
	}
}
