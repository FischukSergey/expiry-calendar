package service_test

import (
	"testing"

	"duekeep/internal/model"
	"duekeep/internal/service"
)

func TestNormalizeCSVMapping(t *testing.T) {
	t.Parallel()
	got, err := service.NormalizeCSVMapping(map[string]string{
		"title": "Name", "kind_slug": "Type", "expires_at": "Until", "attrs.registrar": "Reg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["title"] != "Name" || got["attrs.registrar"] != "Reg" {
		t.Fatalf("%+v", got)
	}
	if _, err := service.NormalizeCSVMapping(nil); err == nil {
		t.Fatal("empty")
	}
	if _, err := service.NormalizeCSVMapping(map[string]string{"unknown": "X"}); err == nil {
		t.Fatal("unknown field")
	}
	if _, err := service.NormalizeCSVMapping(map[string]string{"attrs.": "X"}); err == nil {
		t.Fatal("empty attr key")
	}
}

func TestParseCSVAttr(t *testing.T) {
	t.Parallel()
	v, err := service.ParseCSVAttr(model.AttrString, "  abc ")
	if err != nil || v != "abc" {
		t.Fatalf("string %v %v", v, err)
	}
	v, err = service.ParseCSVAttr(model.AttrNumber, "12.5")
	if err != nil || v.(float64) != 12.5 {
		t.Fatalf("number %v %v", v, err)
	}
	v, err = service.ParseCSVAttr(model.AttrBoolean, "yes")
	flag, ok := v.(bool)
	if err != nil || !ok || !flag {
		t.Fatalf("bool %v %v", v, err)
	}
	if _, err := service.ParseCSVAttr(model.AttrNumber, "x"); err == nil {
		t.Fatal("bad number")
	}
}
