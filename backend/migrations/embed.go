// Package migrations хранит SQL-миграции goose во встроенной FS.
package migrations

import "embed"

// FS — содержимое каталога migrations для goose.SetBaseFS.
//
//go:embed *.sql
var FS embed.FS
