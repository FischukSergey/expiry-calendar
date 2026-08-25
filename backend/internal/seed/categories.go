package seed

type categorySeed struct {
	id        string
	parentID  string
	name      string
	sortOrder int
}

const (
	catIT        = "44444444-4444-4444-4444-444444444401"
	catFinance   = "44444444-4444-4444-4444-444444444402"
	catProperty  = "44444444-4444-4444-4444-444444444403"
	catDocuments = "44444444-4444-4444-4444-444444444404"
	catTransport = "44444444-4444-4444-4444-444444444405"
)

// Корни + второй уровень (≥ 10 строк). Глубина 2, цикл невозможен.
var categorySeeds = []categorySeed{
	{id: catIT, name: "IT", sortOrder: 0},
	{id: catFinance, name: "Финансы", sortOrder: 1},
	{id: catProperty, name: "Имущество", sortOrder: 2},
	{id: catDocuments, name: "Документы", sortOrder: 3},
	{id: catTransport, name: "Транспорт", sortOrder: 4},
	{id: "44444444-4444-4444-4444-444444444411", parentID: catIT, name: "Домены", sortOrder: 0},
	{id: "44444444-4444-4444-4444-444444444412", parentID: catIT, name: "Подписки", sortOrder: 1},
	{id: "44444444-4444-4444-4444-444444444413", parentID: catIT, name: "Лицензии", sortOrder: 2},
	{id: "44444444-4444-4444-4444-444444444421", parentID: catFinance, name: "Налоги", sortOrder: 0},
	{id: "44444444-4444-4444-4444-444444444431", parentID: catProperty, name: "Аренда", sortOrder: 0},
	{id: "44444444-4444-4444-4444-444444444441", parentID: catDocuments, name: "Договоры", sortOrder: 0},
	{id: "44444444-4444-4444-4444-444444444442", parentID: catDocuments, name: "Страховки", sortOrder: 1},
	{id: "44444444-4444-4444-4444-444444444451", parentID: catTransport, name: "Авто", sortOrder: 0},
}

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
