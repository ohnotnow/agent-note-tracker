package ant

import (
	"database/sql"
	"fmt"
)

// SchemaVersion is the schema version this binary expects after Migrate runs.
const SchemaVersion = 1

// migrations is a forward-only list of DDL scripts. Index i takes the database
// from version i-1 to version i. Index 0 is an unused placeholder so that
// migrations[v] is the script that produces version v.
//
// Never edit a published migration in place. To change the schema, append a
// new entry. The current binary will stop at len(migrations)-1.
var migrations = []string{
	"",
	// v1: entries, project_config, indexes.
	`CREATE TABLE entries (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id   TEXT    NOT NULL UNIQUE,
		kind        TEXT    NOT NULL DEFAULT 'note',
		title       TEXT,
		body        TEXT    NOT NULL,
		issue_id    TEXT,
		created_at  TEXT    NOT NULL,
		updated_at  TEXT    NOT NULL
	);
	CREATE INDEX idx_entries_kind       ON entries(kind);
	CREATE INDEX idx_entries_issue_id   ON entries(issue_id);
	CREATE INDEX idx_entries_created_at ON entries(created_at);
	CREATE TABLE project_config (
		id     INTEGER PRIMARY KEY CHECK (id = 1),
		prefix TEXT    NOT NULL
	);`,
}

// Migrate brings the database to the latest schema version. Safe to call
// repeatedly: already-applied migrations are skipped.
func (s *Store) Migrate() error {
	if err := s.ensureSchemaVersionTable(); err != nil {
		return fmt.Errorf("ensure schema_version: %w", err)
	}
	current, err := s.SchemaVersion()
	if err != nil {
		return fmt.Errorf("read schema_version: %w", err)
	}
	for v := current + 1; v < len(migrations); v++ {
		if err := s.applyMigration(v); err != nil {
			return fmt.Errorf("apply migration v%d: %w", v, err)
		}
	}
	return nil
}

// SchemaVersion returns the highest applied migration version, or 0 if none
// have been applied yet.
func (s *Store) SchemaVersion() (int, error) {
	var v sql.NullInt64
	err := s.DB.QueryRow(`SELECT MAX(version) FROM schema_version`).Scan(&v)
	if err != nil {
		return 0, err
	}
	if !v.Valid {
		return 0, nil
	}
	return int(v.Int64), nil
}

func (s *Store) ensureSchemaVersionTable() error {
	_, err := s.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY)`)
	return err
}

func (s *Store) applyMigration(version int) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(migrations[version]); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO schema_version (version) VALUES (?)`, version); err != nil {
		return err
	}
	return tx.Commit()
}
