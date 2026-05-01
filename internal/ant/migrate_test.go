package ant

import "testing"

func TestMigrate_FreshDB(t *testing.T) {
	s := openMemoryStore(t)

	if err := s.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	for _, name := range []string{"entries", "project_config", "schema_version"} {
		if !tableExists(t, s, name) {
			t.Errorf("expected table %q to exist after Migrate", name)
		}
	}

	for _, name := range []string{"idx_entries_kind", "idx_entries_issue_id", "idx_entries_created_at"} {
		if !indexExists(t, s, name) {
			t.Errorf("expected index %q to exist after Migrate", name)
		}
	}

	v, err := s.SchemaVersion()
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if v != SchemaVersion {
		t.Errorf("schema version = %d, want %d", v, SchemaVersion)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	s := openMemoryStore(t)

	if err := s.Migrate(); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := s.Migrate(); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	var count int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM schema_version`).Scan(&count); err != nil {
		t.Fatalf("count schema_version: %v", err)
	}
	if count != SchemaVersion {
		t.Errorf("schema_version row count = %d, want %d", count, SchemaVersion)
	}
}

func TestMigrate_EntryRoundtrip(t *testing.T) {
	s := openMemoryStore(t)
	if err := s.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	_, err := s.DB.Exec(`INSERT INTO entries (public_id, kind, title, body, issue_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"ant-test1", "note", "Hello", "World", "ait-AbCdE", "2026-05-01T00:00:00Z", "2026-05-01T00:00:00Z")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	var got struct {
		PublicID string
		Kind     string
		Body     string
	}
	row := s.DB.QueryRow(`SELECT public_id, kind, body FROM entries WHERE public_id = ?`, "ant-test1")
	if err := row.Scan(&got.PublicID, &got.Kind, &got.Body); err != nil {
		t.Fatalf("select: %v", err)
	}
	if got.PublicID != "ant-test1" || got.Kind != "note" || got.Body != "World" {
		t.Errorf("unexpected row: %+v", got)
	}
}

// --- helpers --------------------------------------------------------------

func openMemoryStore(t *testing.T) *Store {
	t.Helper()
	s, err := OpenStore(":memory:")
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func tableExists(t *testing.T, s *Store, name string) bool {
	t.Helper()
	var n int
	err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name,
	).Scan(&n)
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	return n > 0
}

func indexExists(t *testing.T, s *Store, name string) bool {
	t.Helper()
	var n int
	err := s.DB.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?`, name,
	).Scan(&n)
	if err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	return n > 0
}
