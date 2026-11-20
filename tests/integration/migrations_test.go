//go:build integration

package integration

import "testing"

func TestMigrationsCreateCitextExtension(t *testing.T) {
	if !hasExtension(t, "citext") {
		t.Error("citext extension missing after migrations")
	}
}

func TestMigrationsDownAndUpRestoreSameState(t *testing.T) {
	ctx := t.Context()
	versionBefore := dbVersion(t)
	schemaBefore := schemaSnapshot(t)

	if _, err := migrator.DownTo(ctx, 0); err != nil {
		t.Fatalf("DownTo(0) = %v", err)
	}
	if hasExtension(t, "citext") {
		t.Error("citext extension still present after down")
	}

	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("Up() = %v", err)
	}
	if got := dbVersion(t); got != versionBefore {
		t.Errorf("version after down/up = %d, before = %d", got, versionBefore)
	}
	if got := schemaSnapshot(t); got != schemaBefore {
		t.Errorf("schema after down/up:\n%s\nbefore:\n%s", got, schemaBefore)
	}
}

func hasExtension(t *testing.T, name string) bool {
	t.Helper()
	var exists bool
	err := db.QueryRowContext(t.Context(), "SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = $1)", name).Scan(&exists)
	if err != nil {
		t.Fatalf("query extension %s: %v", name, err)
	}
	return exists
}

func dbVersion(t *testing.T) int64 {
	t.Helper()
	version, err := migrator.GetDBVersion(t.Context())
	if err != nil {
		t.Fatalf("GetDBVersion() = %v", err)
	}
	return version
}

func schemaSnapshot(t *testing.T) string {
	t.Helper()
	const query = `
		SELECT coalesce(string_agg(item, E'\n' ORDER BY item), '') FROM (
			SELECT 'extension ' || extname FROM pg_extension
			UNION ALL
			SELECT format('column %s.%s %s', table_name, column_name, data_type)
			FROM information_schema.columns WHERE table_schema = 'public'
		) AS schema_items(item)`
	var snapshot string
	if err := db.QueryRowContext(t.Context(), query).Scan(&snapshot); err != nil {
		t.Fatalf("list schema: %v", err)
	}
	return snapshot
}
