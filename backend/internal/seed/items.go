package seed

import (
	"encoding/json"
	"fmt"
	"time"
)

// Стабильные UUID первых демо-записей: тесты и карточки опираются на одни id.
const (
	itemRentID         = "55555555-5555-5555-5555-555555555501"
	itemSubscriptionID = "55555555-5555-5555-5555-555555555502"
	itemDomainID       = "55555555-5555-5555-5555-555555555503"
	itemInsuranceID    = "55555555-5555-5555-5555-555555555504"
)

// itemSeed — строка items. Даты считаются от today; status пустой — StatusAtWrite.
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
	status      string
	attrs       map[string]any
}

func intPtr(v int) *int { return &v }

func itemID(n int) string {
	return fmt.Sprintf("55555555-5555-5555-5555-5555555555%02d", n)
}

// it собирает запись: старт за год до обязательства, порог 30 дней, валюта RUB yearly.
func it(n int, title, kind, cat string, expire int) itemSeed {
	return itemSeed{
		id:         itemID(n),
		title:      title,
		kindSlug:   kind,
		categoryID: cat,
		currency:   currencyRUB,
		billing:    billingYearly,
		startDays:  intPtr(expire - 365),
		expireDays: expire,
		notifyDays: 30,
	}
}

func (s itemSeed) money(cost int, cur, bill string) itemSeed {
	s.cost = cost
	s.currency = cur
	s.billing = bill
	return s
}

func (s itemSeed) from(vendor string) itemSeed         { s.vendor = vendor; return s }
func (s itemSeed) extra(attrs map[string]any) itemSeed { s.attrs = attrs; return s }
func (s itemSeed) desc(v string) itemSeed              { s.description = v; return s }
func (s itemSeed) tagged(v ...string) itemSeed         { s.tags = v; return s }
func (s itemSeed) link(v string) itemSeed              { s.url = v; return s }
func (s itemSeed) hint(v string) itemSeed              { s.account = v; return s }
func (s itemSeed) cancelled() itemSeed                 { s.status = statusCancelled; return s }
func (s itemSeed) archived() itemSeed                  { s.status = statusArchived; return s }

// itemSeeds — каталог FUNCTIONAL: ≥50 записей, даты от Clock.Today.
func itemSeeds() []itemSeed {
	return []itemSeed{
		it(1, "Офис на Тверской", slugRent, catRent, 20).
			money(85000, currencyRUB, billingMonthly).from("ООО Простор").
			extra(map[string]any{attrLandlord: "ООО Простор", attrAddress: "Тверская, 7"}).
			desc("Аренда кабинета, продление за месяц").
			tagged("офис", "москва").link("https://prostor.example/lease").hint("договор А-104"),
		it(2, "GitHub Team", slugSubscription, catSubs, 120).
			money(44, currencyUSD, billingMonthly).from("GitHub").
			extra(map[string]any{attrSeats: 5, attrAutoRenew: true}).
			desc("Подписка на организацию").
			tagged(tagDev, "saas").link("https://github.com").hint("duekeep"),
		it(3, "duekeep.ru", slugDomain, catDomains, -10).
			money(890, currencyRUB, billingYearly).from(vendorRegRu).
			extra(map[string]any{attrRegistrar: vendorRegRu, attrAutoRenew: false}).
			desc("Основной домен").tagged(tagDNS).link("https://duekeep.ru"),
		it(4, "ОСАГО Lada", slugInsurance, catInsure, 12).
			money(12500, currencyRUB, billingYearly).from(vendorIngus).
			extra(map[string]any{attrPolicyNo: "XXX-001", attrInsurer: vendorIngus}).
			desc("Полис на рабочий авто").tagged(tagAuto).hint("полис ХХХ"),
		it(5, "Cursor Pro", slugSubscription, catSubs, 7).
			money(20, currencyUSD, billingMonthly).from("Anysphere").
			extra(map[string]any{attrSeats: 1, attrAutoRenew: true}).
			tagged(tagDev, "ai").link("https://cursor.com"),
		it(6, "iCloud+", slugSubscription, catSubs, 3).
			money(299, currencyRUB, billingMonthly).from("Apple").
			extra(map[string]any{attrSeats: 5, attrAutoRenew: true}).
			tagged(tagCloud).link("https://icloud.com"),
		it(7, "duekeep.io", slugDomain, catDomains, 14).
			money(12, currencyUSD, billingYearly).from(vendorCF).
			extra(map[string]any{attrRegistrar: vendorCF, attrAutoRenew: true}).
			tagged(tagDNS).link("https://duekeep.io"),
		it(8, "Паркинг у дома", slugRent, catRent, 28).
			money(4500, currencyRUB, billingMonthly).from(vendorTSJ).
			extra(map[string]any{attrLandlord: vendorTSJ, attrAddress: "Лесная, 12, м/м 7"}).
			tagged(tagHome),
		it(9, "ДМС семья", slugInsurance, catInsure, 18).
			money(48000, currencyRUB, billingYearly).from("СОГАЗ").
			extra(map[string]any{attrPolicyNo: "DMS-4401", attrInsurer: "СОГАЗ"}).
			tagged("здоровье"),
		it(10, "JetBrains All Products", slugLicense, catLicenses, 10).
			money(249, currencyUSD, billingYearly).from("JetBrains").
			extra(map[string]any{attrLicenseKey: "JB-SEED-10", attrSeats: 2}).
			tagged(tagDev).link("https://account.jetbrains.com"),
		it(11, "ТО Lada", slugVehicle, catAuto, 5).
			money(18000, currencyRUB, billingYearly).from("Дилер Lada").
			extra(map[string]any{attrVIN: vinLadaWork, attrPlate: plateLadaWork}).
			tagged(tagAuto, "то"),
		it(12, "НДС 3 квартал", slugTax, catTaxes, 21).
			money(0, currencyRUB, billingOneTime).from(vendorFNS).
			extra(map[string]any{attrTaxAuth: taxOfficeSeven, attrPeriod: "2026-Q3"}).
			tagged(tagTax),
		it(13, "VPS Timeweb", slugOther, catIT, 9).
			money(990, currencyRUB, billingMonthly).from("Timeweb").
			tagged("хостинг").link("https://timeweb.cloud"),
		it(14, "Яндекс 360", slugSubscription, catSubs, -5).
			money(1699, currencyRUB, billingMonthly).from("Яндекс").
			extra(map[string]any{attrSeats: 3, attrAutoRenew: false}).
			tagged(tagCloud),
		it(15, "Клининг офиса", slugContract, catContracts, -40).
			money(22000, currencyRUB, billingMonthly).from("Чистота").
			extra(map[string]any{attrParty: "ООО Чистота", attrContractNo: "CL-19"}).
			tagged("офис"),
		it(16, "Транспортный налог", slugTax, catTaxes, -15).
			money(8400, currencyRUB, billingYearly).from(vendorFNS).
			extra(map[string]any{attrTaxAuth: taxOfficeSeven, attrPeriod: "2025"}).
			tagged(tagTax, tagAuto),
		it(17, "example.com", slugDomain, catDomains, -90).
			money(14, currencyUSD, billingYearly).from(vendorCheap).
			extra(map[string]any{attrRegistrar: vendorCheap, attrAutoRenew: false}).
			tagged(tagDNS),
		it(18, "Adobe Creative Cloud", slugLicense, catLicenses, -25).
			money(60, currencyUSD, billingMonthly).from("Adobe").
			extra(map[string]any{attrLicenseKey: "AD-SEED-18", attrSeats: 1}).
			tagged("дизайн"),
		it(19, "ChatGPT Team", slugSubscription, catSubs, 0).
			money(25, currencyUSD, billingMonthly).from("OpenAI").
			extra(map[string]any{attrSeats: 2, attrAutoRenew: true}).
			tagged("ai"),
		it(20, "Figma Professional", slugSubscription, catSubs, 45).
			money(15, currencyUSD, billingMonthly).from("Figma").
			extra(map[string]any{attrSeats: 2, attrAutoRenew: true}).
			tagged("дизайн"),
		it(21, "Notion Plus", slugSubscription, catSubs, 60).
			money(10, currencyUSD, billingMonthly).from("Notion").
			extra(map[string]any{attrSeats: 4, attrAutoRenew: true}).
			tagged("docs"),
		it(22, "1Password Families", slugSubscription, catSubs, 90).
			money(20, currencyUSD, billingMonthly).from("1Password").
			extra(map[string]any{attrSeats: 5, attrAutoRenew: true}).
			tagged("безопасность"),
		it(23, "Cloudflare Pro", slugSubscription, catSubs, 150).
			money(20, currencyUSD, billingMonthly).from(vendorCF).
			extra(map[string]any{attrSeats: 1, attrAutoRenew: true}).
			tagged(tagDNS),
		it(24, "Vercel Pro", slugSubscription, catSubs, 180).
			money(20, currencyUSD, billingMonthly).from("Vercel").
			extra(map[string]any{attrSeats: 1, attrAutoRenew: true}).
			tagged("хостинг"),
		it(25, "Linear", slugSubscription, catSubs, 210).
			money(8, currencyUSD, billingMonthly).from("Linear").
			extra(map[string]any{attrSeats: 3, attrAutoRenew: true}).
			tagged(tagDev),
		it(26, "duekeep.dev", slugDomain, catDomains, 75).
			money(990, currencyRUB, billingYearly).from(vendorRegRu).
			extra(map[string]any{attrRegistrar: vendorRegRu, attrAutoRenew: true}).
			tagged(tagDNS),
		it(27, "семья.рф", slugDomain, catDomains, 140).
			money(450, currencyRUB, billingYearly).from(vendorRegRu).
			extra(map[string]any{attrRegistrar: vendorRegRu, attrAutoRenew: true}).
			tagged(tagDNS),
		it(28, "home-lab.net", slugDomain, catDomains, 240).
			money(11, currencyUSD, billingYearly).from("Porkbun").
			extra(map[string]any{attrRegistrar: "Porkbun", attrAutoRenew: true}).
			tagged(tagDNS),
		it(29, "old-blog.ru", slugDomain, catDomains, -200).
			money(199, currencyRUB, billingYearly).from("Beget").
			extra(map[string]any{attrRegistrar: "Beget", attrAutoRenew: false}).
			tagged("архив").archived(),
		it(30, "Netflix", slugSubscription, catSubs, 60).
			money(999, currencyRUB, billingMonthly).from("Netflix").
			extra(map[string]any{attrSeats: 1, attrAutoRenew: false}).
			tagged("медиа").cancelled(),
		it(31, "Квартира на Лесной", slugRent, catRent, 100).
			money(72000, currencyRUB, billingMonthly).from("Собственник").
			extra(map[string]any{attrLandlord: "Иванов А.П.", attrAddress: "Лесная, 12, кв. 41"}).
			tagged(tagHome).hint("договор найма"),
		it(32, "Кладовка", slugRent, catRent, 165).
			money(2500, currencyRUB, billingMonthly).from(vendorTSJ).
			extra(map[string]any{attrLandlord: vendorTSJ, attrAddress: "Лесная, 12, кл. 3"}).
			tagged(tagHome),
		it(33, "Договор хостинга", slugContract, catContracts, 200).
			money(0, currencyRUB, billingYearly).from("Timeweb").
			extra(map[string]any{attrParty: "Timeweb", attrContractNo: "TW-8801"}).
			tagged("it"),
		it(34, "Договор с бухгалтером", slugContract, catContracts, 35).
			money(15000, currencyRUB, billingMonthly).from("ИП Смирнова").
			extra(map[string]any{attrParty: "ИП Смирнова", attrContractNo: "ACC-3"}).
			tagged("финансы"),
		it(35, "NDA подрядчик", slugContract, catContracts, 270).
			money(0, currencyRUB, billingOneTime).from("ООО Код").
			extra(map[string]any{attrParty: "ООО Код", attrContractNo: "NDA-12"}).
			tagged("юр"),
		it(36, "Интернет дома", slugContract, catContracts, 310).
			money(890, currencyRUB, billingMonthly).from(vendorMTS).
			extra(map[string]any{attrParty: vendorMTS, attrContractNo: "INET-441"}).
			tagged(tagHome),
		it(37, "КАСКО Lada", slugInsurance, catInsure, 250).
			money(34000, currencyRUB, billingYearly).from(vendorIngus).
			extra(map[string]any{attrPolicyNo: "KASKO-77", attrInsurer: vendorIngus}).
			tagged(tagAuto),
		it(38, "Страховка квартиры", slugInsurance, catInsure, 330).
			money(6200, currencyRUB, billingYearly).from("АльфаСтрахование").
			extra(map[string]any{attrPolicyNo: "FLAT-9", attrInsurer: "АльфаСтрахование"}).
			tagged(tagHome),
		it(39, "Windows 11 Pro", slugLicense, catLicenses, 400).
			money(0, currencyRUB, billingOneTime).from("Microsoft").
			extra(map[string]any{attrLicenseKey: "XXXXX-WIN11", attrSeats: 1}).
			tagged("ос"),
		it(40, "1С Бухгалтерия", slugLicense, catLicenses, 280).
			money(13000, currencyRUB, billingYearly).from("1С").
			extra(map[string]any{attrLicenseKey: "1C-SEED-40", attrSeats: 1}).
			tagged("учёт"),
		it(41, "macOS App Store", slugLicense, catLicenses, 360).
			money(0, currencyUSD, billingYearly).from("Apple").
			extra(map[string]any{attrLicenseKey: "MAS-SEED", attrSeats: 2}).
			tagged("ос"),
		it(42, "УСН декларация", slugTax, catTaxes, 300).
			money(0, currencyRUB, billingYearly).from(vendorFNS).
			extra(map[string]any{attrTaxAuth: taxOfficeSeven, attrPeriod: "2026"}).
			tagged(tagTax),
		it(43, "Имущественный налог", slugTax, catTaxes, 340).
			money(12600, currencyRUB, billingYearly).from(vendorFNS).
			extra(map[string]any{attrTaxAuth: taxOfficeSeven, attrPeriod: "2026"}).
			tagged(tagTax, tagHome),
		it(44, "ОСАГО второй авто", slugVehicle, catAuto, 80).
			money(9800, currencyRUB, billingYearly).from("РЕСО").
			extra(map[string]any{attrVIN: "XTA21099050987654", attrPlate: "В456ОР777"}).
			tagged(tagAuto),
		it(45, "Шиномонтаж сезон", slugVehicle, catAuto, 190).
			money(4000, currencyRUB, billingYearly).from("Колесо").
			extra(map[string]any{attrVIN: vinLadaWork, attrPlate: plateLadaWork}).
			tagged(tagAuto),
		it(46, "Диагностическая карта", slugVehicle, catAuto, 220).
			money(1500, currencyRUB, billingYearly).from("СТО Север").
			extra(map[string]any{attrVIN: vinLadaWork, attrPlate: plateLadaWork}).
			tagged(tagAuto),
		it(47, "Паспорт загран", slugOther, catDocuments, 380).
			money(0, currencyRUB, billingOneTime).from("МВД").
			tagged("документы"),
		it(48, "Абонемент спорт", slugOther, catProperty, 55).
			money(8900, currencyRUB, billingMonthly).from("World Class").
			tagged("здоровье"),
		it(49, "Хранение в облаке Selectel", slugOther, catIT, 130).
			money(490, currencyRUB, billingMonthly).from("Selectel").
			tagged(tagCloud),
		it(50, "Домофон SIM", slugOther, catProperty, 170).
			money(150, currencyRUB, billingMonthly).from(vendorMTS).
			tagged(tagHome),
		it(51, "Подарочный сертификат", slugOther, catDocuments, 40).
			money(5000, currencyRUB, billingOneTime).from("Ozon").
			tagged("покупки"),
		it(52, "Резервный домен duekeep.com", slugDomain, catDomains, 320).
			money(16, currencyUSD, billingYearly).from(vendorCheap).
			extra(map[string]any{attrRegistrar: vendorCheap, attrAutoRenew: true}).
			tagged(tagDNS),
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

func itemComputedStatus(today time.Time, it itemSeed) string {
	if it.status == statusCancelled || it.status == statusArchived {
		return it.status
	}
	_, expires := itemDates(today, it.startDays, it.expireDays)
	return StatusAtWrite(today, expires, it.notifyDays)
}
