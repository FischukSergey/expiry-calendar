package service

// Поля details в 422. Одна константа на ключ — иначе goconst по пакету.
const (
	fieldName      = "name"
	fieldParentID  = "parent_id"
	fieldKey       = "key"
	detailRequired = "required"
	detailMinZero  = ">= 0"
	detailUUID     = "uuid"
	msgInvalidName = "invalid name"
)
