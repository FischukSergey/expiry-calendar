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
	got, err = service.NormalizeCSVMapping(map[string]string{
		"title": "ColTitle", "kind_slug": "ColKind", "expires_at": "ColExp",
		"notify_before_days": "Days", "status": "St",
	})
	if err != nil || got["notify_before_days"] != "Days" || got["status"] != "St" {
		t.Fatalf("notify/status %+v %v", got, err)
	}
}

func TestParseCSVNotify(t *testing.T) {
	t.Parallel()
	days, off, err := service.ParseCSVNotify("", false)
	if err != nil || off || days != model.DefaultNotifyDays {
		t.Fatalf("default %d %v %v", days, off, err)
	}
	for _, raw := range []string{"", "off", "OFF", "-"} {
		_, off, err = service.ParseCSVNotify(raw, true)
		if err != nil || !off {
			t.Fatalf("%q → off=%v err=%v", raw, off, err)
		}
	}
	days, off, err = service.ParseCSVNotify("7", true)
	if err != nil || off || days != 7 {
		t.Fatalf("7 → %d %v %v", days, off, err)
	}
	if _, _, err = service.ParseCSVNotify("x", true); err == nil {
		t.Fatal("bad number")
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
