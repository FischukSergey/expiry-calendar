package seed

type kindSeed struct {
	id         string
	slug       string
	name       string
	color      string
	attrSchema []attrField
}

type attrField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

func field(key, label, typ string) attrField {
	return attrField{Key: key, Label: label, Type: typ, Required: false}
}

// Стабильные UUID, чтобы повторный seed и тесты опирались на одни id.
var kindSeeds = []kindSeed{
	{
		id:    "33333333-3333-3333-3333-333333333301",
		slug:  "domain",
		name:  "Домен",
		color: "#3B82F6",
		attrSchema: []attrField{
			field("registrar", "Регистратор", "string"),
			field("auto_renew", "Автопродление", "boolean"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333302",
		slug:  "subscription",
		name:  "Подписки",
		color: "#8B5CF6",
		attrSchema: []attrField{
			field("seats", "Места", "number"),
			field("auto_renew", "Автопродление", "boolean"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333303",
		slug:  "rent",
		name:  "Аренда",
		color: "#F59E0B",
		attrSchema: []attrField{
			field("landlord", "Арендодатель", "string"),
			field("address", "Адрес", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333304",
		slug:  "contract",
		name:  "Договор",
		color: "#6366F1",
		attrSchema: []attrField{
			field("counterparty", "Контрагент", "string"),
			field("contract_number", "Номер договора", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333305",
		slug:  "insurance",
		name:  "Страховка",
		color: "#10B981",
		attrSchema: []attrField{
			field("policy_number", "Номер полиса", "string"),
			field("insurer", "Страховщик", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333306",
		slug:  "license",
		name:  "Лицензия",
		color: "#EC4899",
		attrSchema: []attrField{
			field("license_key", "Ключ", "string"),
			field("seats", "Места", "number"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333307",
		slug:  "tax",
		name:  "Налог",
		color: "#EF4444",
		attrSchema: []attrField{
			field("tax_authority", "Инспекция", "string"),
			field("period", "Период", "string"),
		},
	},
	{
		id:    "33333333-3333-3333-3333-333333333308",
		slug:  "vehicle",
		name:  "Авто",
		color: "#14B8A6",
		attrSchema: []attrField{
			field("vin", "VIN", "string"),
			field("plate", "Госномер", "string"),
		},
	},
	{
		id:         "33333333-3333-3333-3333-333333333309",
		slug:       "other",
		name:       "Прочее",
		color:      "#6B7280",
		attrSchema: []attrField{},
	},
}
