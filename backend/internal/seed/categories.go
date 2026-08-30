package seed

// categorySeed — узел дерева. Пустой parentID в INSERT становится NULL (корень).
type categorySeed struct {
	id        string
	parentID  string
	name      string
	sortOrder int
}

// Стабильные UUID. Дети ссылаются на корни, не на gen_random_uuid.
const (
	catIT        = "44444444-4444-4444-4444-444444444401"
	catFinance   = "44444444-4444-4444-4444-444444444402"
	catProperty  = "44444444-4444-4444-4444-444444444403"
	catDocuments = "44444444-4444-4444-4444-444444444404"
	catTransport = "44444444-4444-4444-4444-444444444405"
	catDomains   = "44444444-4444-4444-4444-444444444411"
	catSubs      = "44444444-4444-4444-4444-444444444412"
	catLicenses  = "44444444-4444-4444-4444-444444444413"
	catTaxes     = "44444444-4444-4444-4444-444444444421"
	catRent      = "44444444-4444-4444-4444-444444444431"
	catContracts = "44444444-4444-4444-4444-444444444441"
	catInsure    = "44444444-4444-4444-4444-444444444442"
	catAuto      = "44444444-4444-4444-4444-444444444451"
)

// Корни + второй уровень (≥ 10 строк). Глубина 2, цикл невозможен.
var categorySeeds = []categorySeed{
	{id: catIT, name: "IT", sortOrder: 0},
	{id: catFinance, name: "Финансы", sortOrder: 1},
	{id: catProperty, name: "Имущество", sortOrder: 2},
	{id: catDocuments, name: "Документы", sortOrder: 3},
	{id: catTransport, name: "Транспорт", sortOrder: 4},
	{id: catDomains, parentID: catIT, name: "Домены", sortOrder: 0},
	{id: catSubs, parentID: catIT, name: "Подписки", sortOrder: 1},
	{id: catLicenses, parentID: catIT, name: "Лицензии", sortOrder: 2},
	{id: catTaxes, parentID: catFinance, name: "Налоги", sortOrder: 0},
	{id: catRent, parentID: catProperty, name: "Аренда", sortOrder: 0},
	{id: catContracts, parentID: catDocuments, name: "Договоры", sortOrder: 0},
	{id: catInsure, parentID: catDocuments, name: "Страховки", sortOrder: 1},
	{id: catAuto, parentID: catTransport, name: "Авто", sortOrder: 0},
}

// categoryDepth считает уровни от узла к корню (корень = 1). Цикл или дыра → -1.
func categoryDepth(cats []categorySeed, id string) int {
	byID := make(map[string]categorySeed, len(cats))
	for _, c := range cats {
		byID[c.id] = c
	}
	depth := 1
	seen := map[string]struct{}{id: {}}
	cur := byID[id]
	for cur.parentID != "" {
		if _, ok := seen[cur.parentID]; ok {
			return -1
		}
		seen[cur.parentID] = struct{}{}
		parent, ok := byID[cur.parentID]
		if !ok {
			return -1
		}
		depth++
		cur = parent
	}
	return depth
}
