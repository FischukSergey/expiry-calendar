// Package catalog — шаблон категорий для нового пользователя. Это не демо-каталог.
package catalog

// Category — узел шаблона без id. ParentIdx = -1 корень, иначе индекс в слайсе.
type Category struct {
	Name      string
	ParentIdx int
	SortOrder int
}

// DefaultCategories — дерево, которое Register копирует новому владельцу. Без items.
func DefaultCategories() []Category {
	return []Category{
		{Name: "IT", ParentIdx: -1, SortOrder: 0},
		{Name: "Финансы", ParentIdx: -1, SortOrder: 1},
		{Name: "Имущество", ParentIdx: -1, SortOrder: 2},
		{Name: "Документы", ParentIdx: -1, SortOrder: 3},
		{Name: "Транспорт", ParentIdx: -1, SortOrder: 4},
		{Name: "Домены", ParentIdx: 0, SortOrder: 0},
		{Name: "Подписки", ParentIdx: 0, SortOrder: 1},
		{Name: "Лицензии", ParentIdx: 0, SortOrder: 2},
		{Name: "Налоги", ParentIdx: 1, SortOrder: 0},
		{Name: "Аренда", ParentIdx: 2, SortOrder: 0},
		{Name: "Договоры", ParentIdx: 3, SortOrder: 0},
		{Name: "Страховки", ParentIdx: 3, SortOrder: 1},
		{Name: "Авто", ParentIdx: 4, SortOrder: 0},
	}
}
