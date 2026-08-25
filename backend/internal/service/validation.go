package service

// Поля details в 422. Одна константа на ключ — иначе goconst по пакету.
const (
	fieldName      = "name"
	fieldParentID  = "parent_id"
	detailRequired = "required"
	msgInvalidName = "invalid name"
)
