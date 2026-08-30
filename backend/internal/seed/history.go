package seed

import (
	"encoding/json"
	"fmt"
	"time"
)

type renewalSeed struct {
	n         int
	itemN     int
	oldExpire int
	newExpire int
	oldCost   int
	newCost   int
	comment   string
}

type auditSeed struct {
	n      int
	action string
	itemN  int
	renewN int
}

func renewalID(n int) string {
	return fmt.Sprintf("66666666-6666-6666-6666-6666666666%02d", n)
}

func auditID(n int) string {
	return fmt.Sprintf("77777777-7777-7777-7777-7777777777%02d", n)
}

func noteID(itemN int) string {
	return fmt.Sprintf("88888888-8888-8888-8888-8888888888%02d", itemN)
}

func dateOnly(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// renewalSeeds — ≥20 продлений, часть с ростом цены. Даты от today.
func renewalSeeds() []renewalSeed {
	return []renewalSeed{
		{1, 1, -345, 20, 79000, 85000, "индексация аренды"},
		{2, 1, -710, -345, 75000, 79000, "прошлый год"},
		{3, 2, -245, 120, 40, 44, "GitHub +$4"},
		{4, 3, -375, -10, 790, 890, "duekeep.ru 2025"},
		{5, 4, -353, 12, 11000, 12500, "ОСАГО подорожало"},
		{6, 7, -351, 14, 10, 12, "Cloudflare domain"},
		{7, 8, -337, 28, 4000, 4500, "паркинг"},
		{8, 9, -347, 18, 42000, 48000, "ДМС"},
		{9, 10, -355, 10, 229, 249, "JetBrains"},
		{10, 11, -360, 5, 15000, 18000, "ТО"},
		{11, 20, -320, 45, 12, 15, "Figma"},
		{12, 21, -305, 60, 8, 10, "Notion"},
		{13, 22, -275, 90, 18, 20, "1Password"},
		{14, 26, -290, 75, 890, 990, "duekeep.dev"},
		{15, 27, -225, 140, 399, 450, "семья.рф"},
		{16, 31, -265, 100, 68000, 72000, "квартира"},
		{17, 37, -115, 250, 31000, 34000, "КАСКО"},
		{18, 38, -35, 330, 5800, 6200, "квартира страхование"},
		{19, 40, -85, 280, 12000, 13000, "1С"},
		{20, 44, -285, 80, 9000, 9800, "второй ОСАГО"},
		{21, 52, -45, 320, 14, 16, "duekeep.com"},
		{22, 23, -215, 150, 20, 20, "Cloudflare Pro без смены цены"},
	}
}

// auditSeeds — create по первым 16 записям и renew по первым 8 продлениям.
func auditSeeds() []auditSeed {
	out := make([]auditSeed, 0, 24)
	n := 1
	for i := range 16 {
		out = append(out, auditSeed{n: n, action: actionCreate, itemN: i + 1})
		n++
	}
	for i := range 8 {
		out = append(out, auditSeed{n: n, action: actionRenew, itemN: renewalSeeds()[i].itemN, renewN: i + 1})
		n++
	}
	return out
}

func itemByN(n int) (itemSeed, bool) {
	want := itemID(n)
	for _, it := range itemSeeds() {
		if it.id == want {
			return it, true
		}
	}
	return itemSeed{}, false
}

func seedAuditAfter(today time.Time, it itemSeed) ([]byte, error) {
	_, expires := itemDates(today, it.startDays, it.expireDays)
	var cat any
	if it.categoryID != "" {
		cat = it.categoryID
	}
	attrs := it.attrs
	if attrs == nil {
		attrs = map[string]any{}
	}
	return json.Marshal(map[string]any{
		"id":          it.id,
		"title":       it.title,
		"kind_id":     kindIDBySlug(it.kindSlug),
		"category_id": cat,
		"status":      itemComputedStatus(today, it),
		"expires_at":  dateOnly(expires),
		"cost_amount": it.cost,
		"attrs":       attrs,
	})
}
