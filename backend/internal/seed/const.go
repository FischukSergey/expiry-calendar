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

	statusActive    = "active"
	statusExpiring  = "expiring"
	statusExpired   = "expired"
	statusCancelled = "cancelled"
	statusArchived  = "archived"

	actionCreate = "create"
	actionRenew  = "renew"

	attrRegistrar  = "registrar"
	attrAutoRenew  = "auto_renew"
	attrSeats      = "seats"
	attrLandlord   = "landlord"
	attrAddress    = "address"
	attrPolicyNo   = "policy_number"
	attrInsurer    = "insurer"
	attrLicenseKey = "license_key"
	attrTaxAuth    = "tax_authority"
	attrPeriod     = "period"
	attrParty      = "counterparty"
	attrContractNo = "contract_number"
	attrVIN        = "vin"
	attrPlate      = "plate"

	vendorRegRu    = "Reg.ru"
	vendorIngus    = "Ингосстрах"
	vendorTSJ      = "ТСЖ Рассвет"
	vendorCF       = "Cloudflare"
	vendorFNS      = "ФНС"
	vendorMTS      = "МТС"
	vendorCheap    = "Namecheap"
	vinLadaWork    = "XTA21099050123456"
	plateLadaWork  = "А123ВС777"
	taxOfficeSeven = "ИФНС 7"

	tagDev   = "dev"
	tagDNS   = "dns"
	tagHome  = "дом"
	tagAuto  = "авто"
	tagTax   = "налог"
	tagCloud = "облако"
)
