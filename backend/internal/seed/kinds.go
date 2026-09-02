package seed

// kindSeed — строка item_kinds. Идемпотентность по slug, не по id.
type kindSeed struct {
	id         string
	slug       string
	name       string
	color      string
	attrSchema []attrField
}

// attrField — описатель поля формы (не JSON Schema). Type: string|number|boolean.
type attrField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// field собирает необязательный атрибут seed-схемы (required=false).
func field(key, label, typ string) attrField {
	return attrField{Key: key, Label: label, Type: typ, Required: false}
}

// Стабильные UUID, чтобы повторный seed и тесты опирались на одни id.
var kindSeeds = []kindSeed{
	{
		id:    "33333333-3333-3333-3333-333333333301",
		slug:  slugDomain,
		name:  "Домен",
		color: "#3B82F6",
		attrSchema: []attrField{
			field("registrar", "Регистратор", "string"),
			field("auto_renew", "Автопродление", "boolean"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333302",
		slug:  slugSubscription,
		name:  "Подписки",
		color: "#8B5CF6",
		attrSchema: []attrField{
			field("seats", "Места", "number"),
			field("auto_renew", "Автопродление", "boolean"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333303",
		slug:  slugRent,
		name:  "Аренда",
		color: "#F59E0B",
		attrSchema: []attrField{
			field("landlord", "Арендодатель", "string"),
			field("address", "Адрес", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333304",
		slug:  slugContract,
		name:  "Договор",
		color: "#6366F1",
		attrSchema: []attrField{
			field("counterparty", "Контрагент", "string"),
			field("contract_number", "Номер договора", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333305",
		slug:  slugInsurance,
		name:  "Страховка",
		color: "#10B981",
		attrSchema: []attrField{
			field("policy_number", "Номер полиса", "string"),
			field("insurer", "Страховщик", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333306",
		slug:  slugLicense,
		name:  "Лицензия",
		color: "#EC4899",
		attrSchema: []attrField{
			field("license_key", "Ключ", "string"),
			field("seats", "Места", "number"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333307",
		slug:  slugTax,
		name:  "Налог",
		color: "#EF4444",
		attrSchema: []attrField{
			field("tax_authority", "Инспекция", "string"),
			field("period", "Период", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333308",
		slug:  slugVehicle,
		name:  "Авто",
		color: "#14B8A6",
		attrSchema: []attrField{
			field("vin", "VIN", "string"),
			field("plate", "Госномер", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333310",
		slug:  slugMobile,
		name:  "Мобильная связь",
		color: "#0EA5E9",
		attrSchema: []attrField{
			field(attrPhone, "Номер телефона", "string"),
			field(attrOperator, "Оператор", "string"),
		},
	},
	{
		id:         "33333333-3333-3333-3333-333333333309",
		slug:       slugOther,
		name:       "Прочее",
		color:      "#6B7280",
		attrSchema: []attrField{},
	},
}
