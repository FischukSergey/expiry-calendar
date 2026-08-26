package seed

// Slug, валюта, период и статус — одни литералы на пакет (goconst).
const (
	slugDomain       = "domain"
	slugSubscription = "subscription"
	slugRent         = "rent"
	slugContract     = "contract"
	slugInsurance    = "insurance"
	slugLicense      = "license"
	slugTax          = "tax"
	slugVehicle      = "vehicle"
	slugOther        = "other"

	currencyRUB = "RUB"
	currencyUSD = "USD"

	billingOneTime = "one_time"
	billingMonthly = "monthly"
	billingYearly  = "yearly"

	statusActive   = "active"
	statusExpiring = "expiring"
	statusExpired  = "expired"
)
