// Package migrations expõe as migrations SQL do Goose como embed.FS.
package migrations

import "embed"

// FS contém as migrations SQL, para aplicá-las a partir do código (ex.: testes
// de integração).
//
//go:embed *.sql
var FS embed.FS
