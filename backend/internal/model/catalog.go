package model

const (
	// MaxCategoryDepth — корень + двое потомков (корень = 1).
	MaxCategoryDepth = 3
)

// Допустимые type в attr_schema.
const (
	AttrString  = "string"
	AttrNumber  = "number"
	AttrBoolean = "boolean"
)

// AttrField — описатель поля формы, не JSON Schema.
type AttrField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// Kind — строка item_kinds. created_at в API не отдаём.
type Kind struct {
	ID         string      `json:"id"`
	Slug       string      `json:"slug"`
	Name       string      `json:"name"`
	Color      string      `json:"color"`
	AttrSchema []AttrField `json:"attr_schema"`
}

// KindList — GET /kinds без пагинации.
type KindList struct {
	Items []Kind `json:"items"`
}

// Category — узел дерева. Children всегда массив, не null.
type Category struct {
	ID        string     `json:"id"`
	ParentID  *string    `json:"parent_id"`
	Name      string     `json:"name"`
	SortOrder int        `json:"sort_order"`
	Children  []Category `json:"children"`
}

// CategoryList — GET /categories: только корни, дети вложены.
type CategoryList struct {
	Items []Category `json:"items"`
}

// KindPatch — частичное обновление. nil = поле не прислали.
type KindPatch struct {
	Slug       *string
	Name       *string
	Color      *string
	AttrSchema *[]AttrField
}

// CategoryPatch — parent_id: SetParent=false не трогаем; true + ParentID=nil — в корни.
type CategoryPatch struct {
	Name      *string
	SortOrder *int
	SetParent bool
	ParentID  *string
}
