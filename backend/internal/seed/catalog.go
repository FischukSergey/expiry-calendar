package seed

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// requiredKindSlugs — девять slug из ARCHITECTURE.md. ssl и warranty сюда не входят.
var requiredKindSlugs = []string{
	slugDomain, slugSubscription, slugRent, slugContract, slugInsurance,
	slugLicense, slugTax, slugVehicle, slugOther,
}

// forbiddenKindSlugs — сознательно не в seed; admin может завести тип позже.
var forbiddenKindSlugs = []string{"ssl", "warranty"}

// CheckCatalog проверяет инварианты seed (пользователи, kinds, categories, items) без БД.
func CheckCatalog() error {
	if len(userSeeds) != 2 {
		return fmt.Errorf("want 2 users, got %d", len(userSeeds))
	}
	if len(kindSeeds) != len(requiredKindSlugs) {
		return fmt.Errorf("want %d kinds, got %d", len(requiredKindSlugs), len(kindSeeds))
	}

	slugs := make([]string, 0, len(kindSeeds))
	kindIDs := make(map[string]struct{}, len(kindSeeds))
	for _, k := range kindSeeds {
		if _, dup := kindIDs[k.id]; dup {
			return fmt.Errorf("duplicate kind id %s", k.id)
		}
		kindIDs[k.id] = struct{}{}
		slugs = append(slugs, k.slug)
		if err := checkAttrSchema(k.slug, k.attrSchema); err != nil {
			return err
		}
	}
	for _, slug := range requiredKindSlugs {
		if !slices.Contains(slugs, slug) {
			return fmt.Errorf("missing kind %s", slug)
		}
	}
	for _, slug := range forbiddenKindSlugs {
		if slices.Contains(slugs, slug) {
			return fmt.Errorf("forbidden kind %s", slug)
		}
	}

	if len(categorySeeds) < 10 {
		return fmt.Errorf("want at least 10 categories, got %d", len(categorySeeds))
	}

	catIDs := make(map[string]struct{}, len(categorySeeds))
	for _, c := range categorySeeds {
		if _, dup := catIDs[c.id]; dup {
			return fmt.Errorf("duplicate category id %s", c.id)
		}
		catIDs[c.id] = struct{}{}
		depth := categoryDepth(categorySeeds, c.id)
		if depth < 1 {
			return fmt.Errorf("category %s: cycle or missing parent", c.name)
		}
		if depth > 3 {
			return fmt.Errorf("category %s: depth %d > 3", c.name, depth)
		}
	}

	if err := checkItemSeeds(); err != nil {
		return err
	}
	return checkHistorySeeds()
}

// checkItemSeeds: ≥50 записей FUNCTIONAL, уникальные id, kind/category, attrs по схеме.
func checkItemSeeds() error {
	if itemID(1) != itemRentID || itemID(2) != itemSubscriptionID ||
		itemID(3) != itemDomainID || itemID(4) != itemInsuranceID {
		return errors.New("stable item ids drifted")
	}
	items := itemSeeds()
	if len(items) < 50 {
		return fmt.Errorf("want at least 50 items, got %d", len(items))
	}
	ids := make(map[string]struct{}, len(items))
	for _, it := range items {
		if _, dup := ids[it.id]; dup {
			return fmt.Errorf("duplicate item id %s", it.id)
		}
		ids[it.id] = struct{}{}
		if kindIDBySlug(it.kindSlug) == "" {
			return fmt.Errorf("item %s: unknown kind %s", it.title, it.kindSlug)
		}
		if it.categoryID != "" && !categoryExists(it.categoryID) {
			return fmt.Errorf("item %s: unknown category %s", it.title, it.categoryID)
		}
		if it.cost < 0 {
			return fmt.Errorf("item %s: cost_amount must be >= 0", it.title)
		}
		if it.currency == "" || len(it.currency) != 3 {
			return fmt.Errorf("item %s: currency must be ISO 4217", it.title)
		}
		switch it.billing {
		case billingOneTime, billingMonthly, billingYearly:
		default:
			return fmt.Errorf("item %s: bad billing_period %s", it.title, it.billing)
		}
		if err := checkItemAttrs(it); err != nil {
			return err
		}
	}
	today := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	expired, expiring := 0, 0
	for _, it := range items {
		switch itemComputedStatus(today, it) {
		case statusExpired:
			expired++
		case statusExpiring:
			expiring++
		}
	}
	if expired < 5 {
		return fmt.Errorf("want at least 5 expired items, got %d", expired)
	}
	if expiring < 8 {
		return fmt.Errorf("want at least 8 expiring items, got %d", expiring)
	}
	return nil
}

// checkHistorySeeds: ≥20 renewals, ≥15 audit, unread notifications на expired/expiring.
func checkHistorySeeds() error {
	if len(renewalSeeds()) < 20 {
		return fmt.Errorf("want at least 20 renewals, got %d", len(renewalSeeds()))
	}
	for _, r := range renewalSeeds() {
		if _, ok := itemByN(r.itemN); !ok {
			return fmt.Errorf("renewal %d: unknown item %d", r.n, r.itemN)
		}
		if r.oldCost < 0 || r.newCost < 0 {
			return fmt.Errorf("renewal %d: cost must be >= 0", r.n)
		}
	}
	if len(auditSeeds()) < 15 {
		return fmt.Errorf("want at least 15 audit, got %d", len(auditSeeds()))
	}
	for _, a := range auditSeeds() {
		if _, ok := itemByN(a.itemN); !ok {
			return fmt.Errorf("audit %d: unknown item %d", a.n, a.itemN)
		}
		switch a.action {
		case actionCreate, actionRenew:
		default:
			return fmt.Errorf("audit %d: unexpected action %s", a.n, a.action)
		}
	}
	unread := 0
	today := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	for _, it := range itemSeeds() {
		st := itemComputedStatus(today, it)
		if st == statusExpired || st == statusExpiring {
			unread++
		}
	}
	if unread < 1 {
		return errors.New("want unread notifications, got 0")
	}
	return nil
}

func checkItemAttrs(it itemSeed) error {
	var schema []attrField
	for _, k := range kindSeeds {
		if k.slug == it.kindSlug {
			schema = k.attrSchema
			break
		}
	}
	allowed := make(map[string]string, len(schema))
	for _, f := range schema {
		allowed[f.Key] = f.Type
	}
	for key, val := range it.attrs {
		typ, ok := allowed[key]
		if !ok {
			return fmt.Errorf("item %s: extra attr %s", it.title, key)
		}
		if !attrValueMatches(typ, val) {
			return fmt.Errorf("item %s: attr %s want %s", it.title, key, typ)
		}
	}
	return nil
}

func attrValueMatches(typ string, val any) bool {
	switch typ {
	case "string":
		_, ok := val.(string)
		return ok
	case "number":
		switch val.(type) {
		case int, int64, float64:
			return true
		default:
			return false
		}
	case "boolean":
		_, ok := val.(bool)
		return ok
	default:
		return false
	}
}

// checkAttrSchema допускает только string|number|boolean и уникальные key.
func checkAttrSchema(slug string, fields []attrField) error {
	seen := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		if f.Key == "" || f.Label == "" {
			return fmt.Errorf("kind %s: empty attr key/label", slug)
		}
		if _, dup := seen[f.Key]; dup {
			return fmt.Errorf("kind %s: duplicate attr %s", slug, f.Key)
		}
		seen[f.Key] = struct{}{}
		switch f.Type {
		case "string", "number", "boolean":
		default:
			return fmt.Errorf("kind %s: bad attr type %s", slug, f.Type)
		}
	}
	return nil
}
