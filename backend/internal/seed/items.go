package seed

import (
	"encoding/json"
	"time"
)

// Стабильные UUID демо-записей: тесты и повторный seed опираются на одни id.
const (
	itemRentID         = "55555555-5555-5555-5555-555555555501"
	itemSubscriptionID = "55555555-5555-5555-5555-555555555502"
	itemDomainID       = "55555555-5555-5555-5555-555555555503"
	itemInsuranceID    = "55555555-5555-5555-5555-555555555504"
)

const (
	catDomains = "44444444-4444-4444-4444-444444444411"
	catSubs    = "44444444-4444-4444-4444-444444444412"
	catRent    = "44444444-4444-4444-4444-444444444431"
	catInsure  = "44444444-4444-4444-4444-444444444442"
)

// itemSeed — строка items. Даты считаются от today, статус — при вставке.
type itemSeed struct {
	id          string
	title       string
	description string
	kindSlug    string
	categoryID  string
	vendor      string
	tags        []string
	cost        int
	currency    string
	billing     string
	startDays   *int
	expireDays  int
	notifyDays  int
	url         string
	account     string
	attrs       map[string]any
}

func intPtr(v int) *int { return &v }

// itemSeeds — минимум для тестов и демо Sprint 3. Полные 50+ — Sprint 6.
func itemSeeds() []itemSeed {
	return []itemSeed{
		{
			id:          itemRentID,
			title:       "Офис на Тверской",
			description: "Аренда кабинета, продление за месяц",
			kindSlug:    slugRent,
			categoryID:  catRent,
			vendor:      "ООО Простор",
			tags:        []string{"офис", "москва"},
			cost:        85000,
			currency:    currencyRUB,
			billing:     billingMonthly,
			startDays:   intPtr(-200),
			expireDays:  20,
			notifyDays:  30,
			url:         "https://prostor.example/lease",
			account:     "договор А-104",
			attrs: map[string]any{
				"landlord": "ООО Простор",
				"address":  "Тверская, 7",
			},
		},
		{
			id:          itemSubscriptionID,
			title:       "GitHub Team",
			description: "Подписка на организацию",
			kindSlug:    slugSubscription,
			categoryID:  catSubs,
			vendor:      "GitHub",
			tags:        []string{"dev", "saas"},
			cost:        44,
			currency:    currencyUSD,
			billing:     billingMonthly,
			startDays:   intPtr(-90),
			expireDays:  120,
			notifyDays:  30,
			url:         "https://github.com",
			account:     "duekeep",
			attrs: map[string]any{
				"seats":      5,
				"auto_renew": true,
			},
		},
		{
			id:          itemDomainID,
			title:       "duekeep.ru",
			description: "Основной домен",
			kindSlug:    slugDomain,
			categoryID:  catDomains,
			vendor:      "Reg.ru",
			tags:        []string{"dns"},
			cost:        890,
			currency:    currencyRUB,
			billing:     billingYearly,
			startDays:   intPtr(-400),
			expireDays:  -10,
			notifyDays:  30,
			url:         "https://duekeep.ru",
			attrs: map[string]any{
				"registrar":  "Reg.ru",
				"auto_renew": false,
			},
		},
		{
			id:          itemInsuranceID,
			title:       "ОСАГО Lada",
			description: "Полис на рабочий авто",
			kindSlug:    slugInsurance,
			categoryID:  catInsure,
			vendor:      "Ингосстрах",
			tags:        []string{"авто"},
			cost:        12500,
			currency:    currencyRUB,
			billing:     billingYearly,
			startDays:   intPtr(-300),
			expireDays:  12,
			notifyDays:  30,
			url:         "",
			account:     "полис ХХХ",
			attrs: map[string]any{
				"policy_number": "XXX-001",
				"insurer":       "Ингосстрах",
			},
		},
	}
}

func kindIDBySlug(slug string) string {
	for _, k := range kindSeeds {
		if k.slug == slug {
			return k.id
		}
	}
	return ""
}

func categoryExists(id string) bool {
	for _, c := range categorySeeds {
		if c.id == id {
			return true
		}
	}
	return false
}

// StatusAtWrite считает active/expiring/expired. cancelled/archived не сюда.
func StatusAtWrite(today, expires time.Time, notifyDays int) string {
	today = today.UTC().Truncate(24 * time.Hour)
	expires = expires.UTC().Truncate(24 * time.Hour)
	if expires.Before(today) {
		return statusExpired
	}
	until := today.AddDate(0, 0, notifyDays)
	if !expires.After(until) {
		return statusExpiring
	}
	return statusActive
}

func itemDates(today time.Time, startDays *int, expireDays int) (started any, expires time.Time) {
	expires = today.AddDate(0, 0, expireDays)
	if startDays == nil {
		return nil, expires
	}
	return today.AddDate(0, 0, *startDays), expires
}

func marshalAttrs(attrs map[string]any) ([]byte, error) {
	if attrs == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(attrs)
}
