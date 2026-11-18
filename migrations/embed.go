// Package migrations expõe as migrations SQL do Goose como embed.FS.
package migrations

import "embed"

// FS contém as migrations SQL.
//
//go:embed *.sql
var FS embed.FS
